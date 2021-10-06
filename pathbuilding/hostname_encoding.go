package pathbuilding

import (
	"encoding/base32"
	"fmt"
	"github.com/golang/protobuf/proto"
)

func EncodeHostname(testCase *ExplicitTestCase) (string, error) {
	trustGraphIdx := -1
	for idx, trustGraph := range ALL_TRUST_GRAPHS {
		if trustGraph.Name() == testCase.TrustGraph.Name() {
			trustGraphIdx = idx
			break
		}
	}
	if trustGraphIdx == -1 {
		return "", fmt.Errorf("unable to locate index of trust graph with name=%s", testCase.TrustGraph.Name())
	}

	srcNodeIdx := -1
	dstNodeIdx := -1
	for idx, nodeName := range testCase.TrustGraph.NodeNames() {
		if nodeName == testCase.SrcNode {
			srcNodeIdx = idx
		}
		if nodeName == testCase.DstNode {
			dstNodeIdx = idx
		}
	}
	if srcNodeIdx == -1 {
		return "", fmt.Errorf("unable to locate index of node with name=%s", testCase.SrcNode)
	}
	if dstNodeIdx == -1 {
		return "", fmt.Errorf("unable to locate index of node with name=%s", testCase.DstNode)
	}

	invalidEdges := make([]uint32, 0, len(testCase.InvalidEdges))
	for _, invalidEdge := range testCase.InvalidEdges {
		invalidEdgeIdx := -1
		for idx, edge := range testCase.TrustGraph.GetAllEdges() {
			if edge.Source == invalidEdge.Source && edge.Destination == invalidEdge.Destination {
				invalidEdgeIdx = idx
				break
			}
		}
		if invalidEdgeIdx == -1 {
			return "", fmt.Errorf("unable to locate index of edge with src=%s, dst=%s", invalidEdge.Source, invalidEdge.Destination)
		}
		invalidEdges = append(invalidEdges, uint32(invalidEdgeIdx))
	}

	heTestCase := &HostnameEncodedTestCase{
		TrustGraph:    uint32(trustGraphIdx),
		SrcNode:       uint32(srcNodeIdx),
		DstNode:       uint32(dstNodeIdx),
		InvalidEdges:  invalidEdges,
		InvalidReason: uint32(testCase.InvalidReason),
	}

	msgBytes, err := proto.Marshal(heTestCase)
	if err != nil {
		return "", err
	}
	hostname := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(msgBytes)
	return hostname, nil
}

func DecodeHostname(hostname string) (*ExplicitTestCase, error) {
	testCaseBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(hostname)
	if err != nil {
		return nil, fmt.Errorf("invalid servername: %s", hostname)
	}
	heTestCase := new(HostnameEncodedTestCase)
	err = proto.Unmarshal(testCaseBytes, heTestCase)
	if err != nil {
		return nil, err
	}

	trustGraph := ALL_TRUST_GRAPHS[heTestCase.TrustGraph]
	nodeNames := trustGraph.NodeNames()
	edges := trustGraph.GetAllEdges()
	invalidEdges := make([]Edge, 0, len(heTestCase.InvalidEdges))
	for _, invalidEdgeIdx := range heTestCase.InvalidEdges {
		invalidEdges = append(invalidEdges, edges[invalidEdgeIdx])
	}

	return &ExplicitTestCase{
		TrustGraph:    trustGraph,
		SrcNode:       nodeNames[heTestCase.SrcNode],
		DstNode:       nodeNames[heTestCase.DstNode],
		InvalidEdges:  invalidEdges,
		InvalidReason: InvalidReason(heTestCase.InvalidReason),
	}, nil
}
