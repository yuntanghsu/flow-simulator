package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	numEndpoints = 20

	numNodes = 10

	minFlowDurationMilliSeconds    = 500
	minDDoSFlowDurationMilliSecond = 1
	maxFlowDurationMilliSeconds    = 120000

	minPacketCount            = 1
	maxPacketCount            = 10000
	maxDDoSPacketCount        = 20
	defaultNumberOfNormalFlow = 100

	// unit minute
	defaultTimeRange = 10
	// unit milliseconds
	defaultDDoSTimeDuration = 50
	defaultNumOfAttackPods  = 5
	defaultNumAttack        = 1

	alpha = "abcdefghijklmnopqrstuvwxyz"
)

var (
	nameRand *rand.Rand

	numberOfPods     int
	numNormalFlow    int
	numDDoSFlow      int
	timeRange        int
	ddosTimeDuration int
	numAttackPods    int
	numAttack        int
	tcpState         string
)

var (
	destPorts = []int{80, 443, 8080}

	labelKeys   = []string{"app"}
	labelValues = map[string][]string{
		"app": {"appA", "appB", "appC", "appD", "appE"},
	}

	clusterIDs = []string{"7e2e1de2-c85f-476e-ab1a-fce1bf83ee2c", "d29cca64-73df-4618-9c9a-3facff6bef36"}

	endpointIPs, endpointLabels, endpointNames, endpointNamespaces, endpointNodes []string
	endpointPorts                                                                 []int

	_, endpointSubnet, _ = net.ParseCIDR("10.0.0.0/16")
)

func init() {
	nameRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, labelKey := range labelKeys {
		if _, ok := labelValues[labelKey]; !ok {
			panic("label key has no corresponding values")
		}
	}

	ones, bits := endpointSubnet.Mask.Size()
	size := 1 << (bits - ones)
	if numEndpoints > size {
		panic("Subnet not large enough for all endpoints")
	}
}

func generateIP(idx int) net.IP {
	ip := endpointSubnet.IP
	offset := len(ip) - 1
	for {
		if idx == 0 {
			break
		}
		ip[offset] = byte(idx % 256)
		offset -= 1
		idx = idx >> 8
	}
	return ip
}

func generateEndpointIPs(num int) {
	endpointIPs = make([]string, num)
	for idx := 0; idx < num; idx++ {
		ip := generateIP(idx)
		endpointIPs[idx] = ip.String()
	}
}

func generateEndpointPorts(num int) {
	endpointPorts = make([]int, num)
	for idx := 0; idx < num; idx++ {
		endpointPorts[idx] = destPorts[rand.Intn(len(destPorts))]
	}
}

func generateEndpointLabels(num int) {
	endpointLabels = make([]string, num)
	for idx := 0; idx < num; idx++ {
		key := "app"
		values := labelValues[key]
		value := values[rand.Intn(len(values))]
		endpointLabels[idx] = fmt.Sprintf(`'{"%s":"%s"}'`, key, value)
	}

}

func generateEndpointNames(num int) {
	endpointNames = make([]string, num)
	for idx := 0; idx < num; idx++ {
		endpointNames[idx] = fmt.Sprintf("pod-%d", idx)
	}
}

func generateEndpointNamespaces(num int) {
	endpointNamespaces = make([]string, num)
	for idx := 0; idx < num; idx++ {
		endpointNamespaces[idx] = fmt.Sprintf("namespace-%s", string(alpha[idx%26]))
	}
}

func generateEndpointNodes(num int) {
	endpointNodes = make([]string, num)
	for idx := 0; idx < num; idx++ {
		endpointNodes[idx] = fmt.Sprintf("node-%d", rand.Intn(numNodes))
	}
}

type FlowRecord struct {
	FlowStartSeconds                     string `json:"flowStartSeconds"`
	FlowEndSeconds                       string `json:"flowEndSeconds"`
	FlowEndSecondsFromSourceNode         string `json:"flowEndSecondsFromSourceNode"`
	FlowEndSecondsFromDestinationNode    string `json:"flowEndSecondsFromDestinationNode"`
	FlowEndReason                        uint8  `json:"flowEndReason"`
	SourceIP                             string `json:"sourceIP"`
	DestinationIP                        string `json:"destinationIP"`
	SourceTransportPort                  uint16 `json:"sourceTransportPort"`
	DestinationTransportPort             uint16 `json:"destinationTransportPort"`
	ProtocolIdentifier                   uint8  `json:"protocolIdentifier"`
	PacketTotalCount                     uint64 `json:"packetTotalCount"`
	OctetTotalCount                      uint64 `json:"octetTotalCount"`
	PacketDeltaCount                     uint64 `json:"packetDeltaCount"`
	OctetDeltaCount                      uint64 `json:"octetDeltaCount"`
	ReversePacketTotalCount              uint64 `json:"reversePacketTotalCount"`
	ReverseOctetTotalCount               uint64 `json:"reverseOctetTotalCount"`
	ReversePacketDeltaCount              uint64 `json:"reversePacketDeltaCount"`
	ReverseOctetDeltaCount               uint64 `json:"reverseOctetDeltaCount"`
	SourcePodName                        string `json:"sourcePodName"`
	SourcePodNamespace                   string `json:"sourcePodNamespace"`
	SourceNodeName                       string `json:"sourceNodeName"`
	DestinationPodName                   string `json:"destinationPodName"`
	DestinationPodNamespace              string `json:"destinationPodNamespace"`
	DestinationNodeName                  string `json:"destinationNodeName"`
	DestinationClusterIP                 string `json:"destinationClusterIP"`
	DestinationServicePort               uint16 `json:"destinationServicePort"`
	DestinationServicePortName           string `json:"destinationServicePortName"`
	IngressNetworkPolicyName             string `json:"ingressNetworkPolicyName"`
	IngressNetworkPolicyNamespace        string `json:"ingressNetworkPolicyNamespace"`
	IngressNetworkPolicyRuleName         string `json:"ingressNetworkPolicyRuleName"`
	IngressNetworkPolicyRuleAction       uint8  `json:"ingressNetworkPolicyRuleAction"`
	IngressNetworkPolicyType             uint8  `json:"ingressNetworkPolicyType"`
	EgressNetworkPolicyName              string `json:"egressNetworkPolicyName"`
	EgressNetworkPolicyNamespace         string `json:"egressNetworkPolicyNamespace"`
	EgressNetworkPolicyRuleName          string `json:"egressNetworkPolicyRuleName"`
	EgressNetworkPolicyRuleAction        uint8  `json:"egressNetworkPolicyRuleAction"`
	EgressNetworkPolicyType              uint8  `json:"egressNetworkPolicyType"`
	TcpState                             string `json:"tcpState"`
	FlowType                             uint8  `json:"flowType"`
	SourcePodLabels                      string `json:"sourcePodLabels"`
	DestinationPodLabels                 string `json:"destinationPodLabels"`
	Throughput                           uint64 `json:"throughput"`
	ReverseThroughput                    uint64 `json:"reverseThroughput"`
	ThroughputFromSourceNode             uint64 `json:"throughputFromSourceNode"`
	ThroughputFromDestinationNode        uint64 `json:"throughputFromDestinationNode"`
	ReverseThroughputFromSourceNode      uint64 `json:"reverseThroughputFromSourceNode"`
	ReverseThroughputFromDestinationNode uint64 `json:"reverseThroughputFromDestinationNode"`
	ClusterUUID                          string `json:"clusterUUID"`
}

func toTimestamp(t time.Time) string {
	const layout = "2006-01-02 15:04:05.0000 -0700"
	return t.Format(layout)
}

func generateFlowRecord(ddos bool, destinationPod int, expectedTime time.Time) *FlowRecord {
	meanTimeFuture := expectedTime
	var dest int
	var durationMilliSeconds uint64
	var packetCount uint64
	var reversePacketCount uint64
	var sourcePodLabel string
	var octetCount uint64
	var reverseOctetCount uint64
	var tcp string
	if ddos {
		dest = destinationPod
		durationMilliSeconds = uint64(minDDoSFlowDurationMilliSecond + rand.Intn(ddosTimeDuration))
		packetCount = uint64(minPacketCount + 4 + rand.Int63n(maxDDoSPacketCount-minPacketCount))
		reversePacketCount = uint64(minPacketCount + 4 + rand.Int63n(maxDDoSPacketCount-minPacketCount))
		octetCount = packetCount * 50
		reverseOctetCount = reversePacketCount * 50
		if tcpState != "" {
			tcp = tcpState
		}
	} else {
		dest = rand.Intn(numberOfPods)
		durationMilliSeconds = uint64(minFlowDurationMilliSeconds + rand.Int63n(maxFlowDurationMilliSeconds-minFlowDurationMilliSeconds))
		packetCount = uint64(minPacketCount + rand.Int63n(maxPacketCount-minPacketCount))
		reversePacketCount = uint64(minPacketCount + rand.Int63n(maxPacketCount-minPacketCount))
		octetCount = packetCount * 1500
		reverseOctetCount = reversePacketCount * 1500
		tcp = "TIME_WAIT"
	}
	source := (dest + 1 + rand.Intn(numAttackPods)) % numberOfPods
	if ddos {
		sourcePodLabel = fmt.Sprintf(`'{"%s":"%s"}'`, "DDOS", "1")
	} else {
		sourcePodLabel = fmt.Sprintf(`'{"%s":"%s"}'`, "DDOS", "0")
	}
	start := toTimestamp(meanTimeFuture.Add(time.Duration(-durationMilliSeconds) * time.Millisecond / 2))
	end := toTimestamp(meanTimeFuture.Add(time.Duration(durationMilliSeconds) * time.Millisecond / 2))
	throughput := uint64(octetCount * 8 / durationMilliSeconds * 1000)
	reverseThroughput := uint64(reverseOctetCount * 8 / durationMilliSeconds * 1000)
	destinationPodName := endpointNames[dest]
	destinationPodNamespace := endpointNamespaces[dest]
	destinationNodeName := endpointNodes[dest]

	var flowType int
	sameNode := endpointNodes[source] == endpointNodes[dest]
	switch sameNode {
	case true:
		flowType = 1
	case false:
		flowType = 2
	}
	randomExternalFlow := rand.Intn(10000)
	if !ddos && randomExternalFlow%10 == 0 {
		flowType = 3
		destinationPodName = ""
		destinationPodNamespace = ""
		destinationNodeName = ""
	}
	return &FlowRecord{
		FlowStartSeconds:                     start,
		FlowEndSeconds:                       end,
		FlowEndSecondsFromSourceNode:         end,
		FlowEndSecondsFromDestinationNode:    end,
		FlowEndReason:                        3,
		SourceIP:                             endpointIPs[source],
		DestinationIP:                        endpointIPs[dest],
		SourceTransportPort:                  uint16(1 + rand.Intn(65534)),
		DestinationTransportPort:             uint16(endpointPorts[dest]),
		ProtocolIdentifier:                   6, // TCP
		PacketTotalCount:                     packetCount,
		OctetTotalCount:                      octetCount,
		PacketDeltaCount:                     packetCount,
		OctetDeltaCount:                      octetCount,
		ReversePacketTotalCount:              reversePacketCount,
		ReverseOctetTotalCount:               reverseOctetCount,
		ReversePacketDeltaCount:              reversePacketCount,
		ReverseOctetDeltaCount:               reverseOctetCount,
		SourcePodName:                        endpointNames[source],
		SourcePodNamespace:                   endpointNamespaces[source],
		SourceNodeName:                       endpointNodes[source],
		DestinationPodName:                   destinationPodName,
		DestinationPodNamespace:              destinationPodNamespace,
		DestinationNodeName:                  destinationNodeName,
		TcpState:                             tcp,
		FlowType:                             uint8(flowType),
		SourcePodLabels:                      sourcePodLabel,
		DestinationPodLabels:                 endpointLabels[dest],
		Throughput:                           throughput,
		ReverseThroughput:                    reverseThroughput,
		ThroughputFromSourceNode:             throughput,
		ThroughputFromDestinationNode:        throughput,
		ReverseThroughputFromSourceNode:      reverseThroughput,
		ReverseThroughputFromDestinationNode: reverseThroughput,
		ClusterUUID:                          clusterIDs[0],
	}
}

func (r *FlowRecord) AsCSV(w io.Writer) {
	io.WriteString(w, toTimestamp(time.Now().Add(time.Duration(timeRange)*time.Minute+time.Duration(maxFlowDurationMilliSeconds)*time.Millisecond)))
	io.WriteString(w, ",")
	io.WriteString(w, r.FlowStartSeconds)
	io.WriteString(w, ",")
	io.WriteString(w, r.FlowEndSeconds)
	io.WriteString(w, ",")
	io.WriteString(w, r.FlowEndSecondsFromSourceNode)
	io.WriteString(w, ",")
	io.WriteString(w, r.FlowEndSecondsFromDestinationNode)
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.FlowEndReason))
	io.WriteString(w, ",")
	io.WriteString(w, r.SourceIP)
	io.WriteString(w, ",")
	io.WriteString(w, r.DestinationIP)
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.SourceTransportPort))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.DestinationTransportPort))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ProtocolIdentifier))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.PacketTotalCount))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.OctetTotalCount))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.PacketDeltaCount))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.OctetDeltaCount))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ReversePacketTotalCount))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ReverseOctetTotalCount))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ReversePacketDeltaCount))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ReverseOctetDeltaCount))
	io.WriteString(w, ",")
	io.WriteString(w, r.SourcePodName)
	io.WriteString(w, ",")
	io.WriteString(w, r.SourcePodNamespace)
	io.WriteString(w, ",")
	io.WriteString(w, r.SourceNodeName)
	io.WriteString(w, ",")
	io.WriteString(w, r.DestinationPodName)
	io.WriteString(w, ",")
	io.WriteString(w, r.DestinationPodNamespace)
	io.WriteString(w, ",")
	io.WriteString(w, r.DestinationNodeName)
	io.WriteString(w, ",")
	io.WriteString(w, r.DestinationClusterIP)
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.DestinationServicePort))
	io.WriteString(w, ",")
	io.WriteString(w, r.DestinationServicePortName)
	io.WriteString(w, ",")
	io.WriteString(w, r.IngressNetworkPolicyName)
	io.WriteString(w, ",")
	io.WriteString(w, r.IngressNetworkPolicyNamespace)
	io.WriteString(w, ",")
	io.WriteString(w, r.IngressNetworkPolicyRuleName)
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.IngressNetworkPolicyRuleAction))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.IngressNetworkPolicyType))
	io.WriteString(w, ",")
	io.WriteString(w, r.EgressNetworkPolicyName)
	io.WriteString(w, ",")
	io.WriteString(w, r.EgressNetworkPolicyNamespace)
	io.WriteString(w, ",")
	io.WriteString(w, r.EgressNetworkPolicyRuleName)
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.EgressNetworkPolicyRuleAction))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.EgressNetworkPolicyType))
	io.WriteString(w, ",")
	io.WriteString(w, r.TcpState)
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.FlowType))
	io.WriteString(w, ",")
	io.WriteString(w, r.SourcePodLabels)
	io.WriteString(w, ",")
	io.WriteString(w, r.DestinationPodLabels)
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.Throughput))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ReverseThroughput))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ThroughputFromSourceNode))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ThroughputFromDestinationNode))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ReverseThroughputFromSourceNode))
	io.WriteString(w, ",")
	io.WriteString(w, fmt.Sprintf("%d", r.ReverseThroughputFromDestinationNode))
	io.WriteString(w, ",")
	io.WriteString(w, r.ClusterUUID)
}

func writeHeader(w io.Writer) {
	io.WriteString(w, "InsertedTime")
	io.WriteString(w, ",")
	io.WriteString(w, "FlowStartSeconds")
	io.WriteString(w, ",")
	io.WriteString(w, "FlowEndSeconds")
	io.WriteString(w, ",")
	io.WriteString(w, "FlowEndSecondsFromSourceNode")
	io.WriteString(w, ",")
	io.WriteString(w, "FlowEndSecondsFromDestinationNode")
	io.WriteString(w, ",")
	io.WriteString(w, "FlowEndReason")
	io.WriteString(w, ",")
	io.WriteString(w, "SourceIP")
	io.WriteString(w, ",")
	io.WriteString(w, "DestinationIP")
	io.WriteString(w, ",")
	io.WriteString(w, "SourceTransportPort")
	io.WriteString(w, ",")
	io.WriteString(w, "DestinationTransportPort")
	io.WriteString(w, ",")
	io.WriteString(w, "ProtocolIdentifier")
	io.WriteString(w, ",")
	io.WriteString(w, "PacketTotalCount")
	io.WriteString(w, ",")
	io.WriteString(w, "OctetTotalCount")
	io.WriteString(w, ",")
	io.WriteString(w, "PacketDeltaCount")
	io.WriteString(w, ",")
	io.WriteString(w, "OctetDeltaCount")
	io.WriteString(w, ",")
	io.WriteString(w, "ReversePacketTotalCount")
	io.WriteString(w, ",")
	io.WriteString(w, "ReverseOctetTotalCount")
	io.WriteString(w, ",")
	io.WriteString(w, "ReversePacketDeltaCount")
	io.WriteString(w, ",")
	io.WriteString(w, "ReverseOctetDeltaCount")
	io.WriteString(w, ",")
	io.WriteString(w, "SourcePodName")
	io.WriteString(w, ",")
	io.WriteString(w, "SourcePodNamespace")
	io.WriteString(w, ",")
	io.WriteString(w, "SourceNodeName")
	io.WriteString(w, ",")
	io.WriteString(w, "DestinationPodName")
	io.WriteString(w, ",")
	io.WriteString(w, "DestinationPodNamespace")
	io.WriteString(w, ",")
	io.WriteString(w, "DestinationNodeName")
	io.WriteString(w, ",")
	io.WriteString(w, "DestinationClusterIP")
	io.WriteString(w, ",")
	io.WriteString(w, "DestinationServicePort")
	io.WriteString(w, ",")
	io.WriteString(w, "DestinationServicePortName")
	io.WriteString(w, ",")
	io.WriteString(w, "IngressNetworkPolicyName")
	io.WriteString(w, ",")
	io.WriteString(w, "IngressNetworkPolicyNamespace")
	io.WriteString(w, ",")
	io.WriteString(w, "IngressNetworkPolicyRuleName")
	io.WriteString(w, ",")
	io.WriteString(w, "IngressNetworkPolicyRuleAction")
	io.WriteString(w, ",")
	io.WriteString(w, "IngressNetworkPolicyType")
	io.WriteString(w, ",")
	io.WriteString(w, "EgressNetworkPolicyName")
	io.WriteString(w, ",")
	io.WriteString(w, "EgressNetworkPolicyNamespace")
	io.WriteString(w, ",")
	io.WriteString(w, "EgressNetworkPolicyRuleName")
	io.WriteString(w, ",")
	io.WriteString(w, "EgressNetworkPolicyRuleAction")
	io.WriteString(w, ",")
	io.WriteString(w, "EgressNetworkPolicyType")
	io.WriteString(w, ",")
	io.WriteString(w, "TcpState")
	io.WriteString(w, ",")
	io.WriteString(w, "FlowType")
	io.WriteString(w, ",")
	io.WriteString(w, "SourcePodLabels")
	io.WriteString(w, ",")
	io.WriteString(w, "DestinationPodLabels")
	io.WriteString(w, ",")
	io.WriteString(w, "Throughput")
	io.WriteString(w, ",")
	io.WriteString(w, "ReverseThroughput")
	io.WriteString(w, ",")
	io.WriteString(w, "ThroughputFromSourceNode")
	io.WriteString(w, ",")
	io.WriteString(w, "ThroughputFromDestinationNode")
	io.WriteString(w, ",")
	io.WriteString(w, "ReverseThroughputFromSourceNode")
	io.WriteString(w, ",")
	io.WriteString(w, "ReverseThroughputFromDestinationNode")
	io.WriteString(w, ",")
	io.WriteString(w, "ClusterUUID")
}

func generateFlowRecords(writer io.Writer) {
	writeHeader(writer)
	io.WriteString(writer, "\n")
	for idx := 0; idx < numNormalFlow; idx++ {
		record := generateFlowRecord(false, 0, time.Now().Add(time.Duration(rand.Intn(timeRange))*time.Minute))
		record.AsCSV(writer)
		io.WriteString(writer, "\n")
	}
	for att := 0; att < numAttack; att++ {
		dest := rand.Intn(numberOfPods)
		attackTime := time.Now().Add(time.Duration(rand.Intn(timeRange)) * time.Minute)
		for idx := 0; idx < numDDoSFlow; idx++ {
			record := generateFlowRecord(true, dest, attackTime)
			record.AsCSV(writer)
			io.WriteString(writer, "\n")
		}
	}
}

func generateFlowRecordsAsCsvFile() (string, error) {
	f, err := os.CreateTemp("./", "records-*.csv")
	if err != nil {
		return "", err
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	defer writer.Flush()
	generateFlowRecords(writer)
	return f.Name(), nil
}

func localStage(num int) error {
	filePath, err := generateFlowRecordsAsCsvFile()
	if err != nil {
		return err
	}
	fmt.Println(filePath)
	return nil
}

func main() {
	//var configPath string
	flag.IntVar(&numberOfPods, "numberOfPods", 0, "number of Pods in the cluster, default is 20")
	flag.IntVar(&numNormalFlow, "numNormalFlow", 0, "number of normal flow, default is 100")
	flag.IntVar(&numDDoSFlow, "numDDoSFlow", 0, "number of ddos flow for each attach, default is 0")
	flag.IntVar(&timeRange, "timeRange", 0, "time range for the start time, unit is minute, default is 10")
	flag.IntVar(&ddosTimeDuration, "ddosTimeDuration", 0, "maximum time length for the ddos flow, unit is millisecond, default is 50")
	flag.IntVar(&numAttackPods, "numAttackPods", 0, "number of DDoS attack Pods, default is 5")
	flag.IntVar(&numAttack, "numAttack", 0, "number of attack times, defaut is 1")
	flag.StringVar(&tcpState, "tcpState", "", "TCP state for attack flow, default is TIME_WAIT")

	flag.Parse()

	if numberOfPods == 0 {
		numberOfPods = numEndpoints
	}
	if numNormalFlow == 0 {
		numNormalFlow = defaultNumberOfNormalFlow
	}
	if timeRange == 0 {
		timeRange = defaultTimeRange
	}
	if ddosTimeDuration == 0 {
		ddosTimeDuration = defaultDDoSTimeDuration
	}
	if numAttackPods == 0 {
		numAttackPods = defaultNumOfAttackPods
	}
	if numAttack == 0 {
		numAttack = defaultNumAttack
	}

	generateEndpointIPs(numberOfPods)
	generateEndpointPorts(numberOfPods)
	generateEndpointLabels(numberOfPods)
	generateEndpointNames(numberOfPods)
	generateEndpointNamespaces(numberOfPods)
	generateEndpointNodes(numberOfPods)

	err := localStage(numberOfPods)

	if err != nil {
		fmt.Printf("Error when generating flow records, err: %v", err)
		os.Exit(1)
	}
}
