package crackutils

import (
	"CharcoalFire/utils"
	"github.com/Workiva/go-datastructures/queue"
	"github.com/miekg/dns"
	"github.com/projectdiscovery/dnsx/libs/dnsx"
	"github.com/projectdiscovery/retryabledns"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Lcdn = utils.GetSlog("cdncheck")
var wg sync.WaitGroup
var validResolversList []string // 可用dns服务器列表
var noCdnDomains []string       // 未使用cdn的域名列表
var useCdnDomains []string      // 使用cdn的域名列表
var noCdnIps []string           // 未使用cdn的ip列表

func CheckCdn(domainList []string) {

	cdnCnameList, err := utils.ReadLinesFromFile("crackutils/cdn_cname")
	if err != nil {
		Lcdn.Fatal("cdn_cname文件解析失败")
	}
	tempResolversList, err := utils.ReadLinesFromFile("crackutils/resolvers.txt")
	if err != nil {
		Lcdn.Fatal("resolvers文件解析失败")
	}

	queResolvers := queue.New(utils.DnsSearchThread) // dns服务器队列
	for _, resolver := range tempResolversList {
		err := queResolvers.Put(resolver)
		if err != nil {
			continue
		}
	}
	for queResolvers.Len() > 0 {
		wg.Add(1)
		queResolverList, _ := queResolvers.Get(1)
		queResolver := queResolverList[0].(string)
		go FilterValidResolver(queResolver)
	}
	wg.Wait()
	dnsSCn := len(validResolversList)
	if dnsSCn < 20 {
		Lcdn.Fatal("有效的 DNS 解析器数量过少，暂无法执行判断CDN的操作")
		return
	} else {
		Lcdn.Info("已经找到 " + strconv.Itoa(dnsSCn) + " 个可用的DNS解析器，为CDN精准识别保驾护航")
	}

	var DefaultOptions = dnsx.Options{
		BaseResolvers:     validResolversList,
		MaxRetries:        5,
		QuestionTypes:     []uint16{dns.TypeA},
		TraceMaxRecursion: math.MaxUint16,
		Hostsfile:         true,
	}
	// 初始化dnsx客户端
	dnsxClient, _ := dnsx.New(DefaultOptions)

	que := queue.New(utils.CDNCheckThread) // 待识别域名队列

	for _, domain := range domainList {
		err := que.Put(domain)
		if err != nil {
			continue
		}
	}

	for que.Len() > 0 {
		wg.Add(1)
		queDomainList, _ := que.Get(1)
		queDomain := queDomainList[0].(string)
		go CdnCheck(queDomain, cdnCnameList, validResolversList, dnsxClient)
	}
	wg.Wait()

	println("no cdn domain")
	for _, v := range noCdnDomains {
		println(v)
	}
	println("use cdn domain")
	for _, v := range useCdnDomains {
		println(v)
	}
	println("no cdn ips")
	for _, v := range noCdnIps {
		println(v)
	}

}

// FilterValidResolver 筛选有效的 DNS 解析器
func FilterValidResolver(resolver string) {
	defer wg.Done()
	retries := 1 // 重试次数
	tempResolverList := []string{resolver}
	dnsClient, _ := retryabledns.New(tempResolverList, retries)
	dnsResponses, _ := dnsClient.Query("public1.114dns.com", dns.TypeA)
	if In("114.114.114.114", dnsResponses.A) {
		validResolversList = append(validResolversList, resolver)
	}
}

// In 判断字符串在切片中是否包含
func In(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	if index < len(strArray) && strArray[index] == target {
		return true
	}
	return false
}

// CdnCheck 判断目标域名是否使用了CDN
func CdnCheck(domain string, cdnCnameList []string, resolversList []string, dnsxClient *dnsx.DNSX) {
	defer wg.Done()
	// 通过dnsx自带方法识别cdn，主要为根据ip范围判断国外主流cdn厂商
	isCdn, _, _ := dnsxClient.CdnCheck(domain)
	dnsxResult, _ := dnsxClient.QueryOne(domain)
	domainCnameList := dnsxResult.CNAME
	domainAList := dnsxResult.A // 域名的 IPv4 地址记录
	var domainIpList []string

	if len(domainAList) > 0 {
		for _, tempIp := range domainAList {
			if !In(tempIp, resolversList) {
				domainIpList = append(domainIpList, tempIp)
			}
		}
	} else {
		return
	}

	if isCdn {
		useCdnDomains = append(useCdnDomains, domain)
		Lcdn.Debug(domain + " 使用了CDN")
		return
	}

	// 2.无cname但有A记录，直接判定未使用cdn，找到目标直接使用自己的服务器进行内容分发情况
	if len(domainCnameList) == 0 && len(domainIpList) > 0 {
		noCdnDomains = append(noCdnDomains, domain)
		Lcdn.Info(domain + " 未使用CDN")
		noCdnIps = UniqueStrList(append(noCdnIps, domainIpList...))
		return
	} else if len(domainCnameList) > 0 && len(domainIpList) > 0 { // 3.cdn在 cdn cname 列表中包含，直接判定使用cdn
		if InCdnCnameList(domainCnameList, cdnCnameList) {
			useCdnDomains = append(useCdnDomains, domain)
			Lcdn.Debug(domain + " 使用了CDN")
			return
		} else {
			var domainIpPartList []string
			randNums := GenerateRandomNumber(0, len(resolversList), 30)
			for _, num := range randNums {
				resolver := resolversList[num]
				domainIpsWithResolver, err := ResolvDomainIpPart(domain, resolver)
				if err != nil {
					continue
				}
				domainIpPartList = UniqueStrList(append(domainIpPartList, domainIpsWithResolver...))
				if len(domainIpPartList) > 3 { // 不同段ip数量达到4个就跳出循环，避免每个dns服务器都解析增加耗时
					println(resolver)
					break
				}
			}
			// 不同段ip数量达到4个就判定为使用了cdn
			if len(domainIpPartList) > 3 {
				useCdnDomains = append(useCdnDomains, domain)
				Lcdn.Debug(domain + " 使用了CDN")
			} else {
				noCdnDomains = append(noCdnDomains, domain)
				Lcdn.Info(domain + " 未使用CDN")
				noCdnIps = UniqueStrList(append(noCdnIps, domainIpList...))
			}
		}
	}
}

// UniqueStrList 切片去重
func UniqueStrList(strList []string) []string {
	uniqList := make([]string, 0)
	tempMap := make(map[string]bool, len(strList))
	for _, v := range strList {
		if tempMap[v] == false && len(v) > 0 {
			tempMap[v] = true
			uniqList = append(uniqList, v)
		}
	}
	return uniqList
}

// InCdnCnameList 判断是否使用了已知的cname
func InCdnCnameList(domainCnameList []string, cdnCnameList []string) bool {
	inCdnCname := false
	for _, domainCname := range domainCnameList {
		for _, cdnCname := range cdnCnameList {
			if strings.Contains(domainCname, cdnCname) {
				inCdnCname = true
				return inCdnCname
			}
		}
	}
	return inCdnCname
}

func GenerateRandomNumber(start int, end int, count int) []int {
	// 范围检查
	if end < start || (end-start) < count {
		return nil
	}
	nums := make([]int, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		num := r.Intn(end-start) + start
		// 查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}

// ResolvDomainIpPart 获取域名在特定dns上的解析ip，并且以.分割ip，取ip的前三部分，即解析ip为1.1.1.1,最终输出为[1.1.1]，便于判断多个ip是否在相同网段
func ResolvDomainIpPart(domain string, resolver string) ([]string, error) {
	var domainIpsPart []string
	retries := 1
	resolverList := []string{resolver}
	dnsClient, err := retryabledns.New(resolverList, retries)
	if err != nil {
		return domainIpsPart, err
	}
	dnsResponses, err := dnsClient.Query(domain, dns.TypeA)
	if err != nil {
		return domainIpsPart, err
	} else if In(resolver, dnsResponses.A) { // 如果dns的ip出现在查询结果中，判为误报，忽略结果
		// DNS 解析器通常会将域名解析请求转发给其他的 DNS 服务器来获取解析结果。因此，原始的 DNS 解析器的 IP 地址可能会出现在查询结果中
		return domainIpsPart, nil
	}
	ipsList := dnsResponses.A
	if len(ipsList) > 0 {
		for _, ip := range ipsList {
			ipParts := strings.Split(ip, ".")
			ipSplit := ipParts[0] + "." + ipParts[1] + "." + ipParts[2]
			domainIpsPart = UniqueStrList(append(domainIpsPart, ipSplit))
		}
	}
	return domainIpsPart, nil
}
