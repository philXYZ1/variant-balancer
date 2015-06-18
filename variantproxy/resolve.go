package variantproxy

import (
	"net/http"
)

func (p *Proxy) ResolveNode(incomingRequest *http.Request) (n *Node, sessionId string) {
	availableNodes := []*Node{}
	sessionId = ""
	for _, node := range p.Nodes {
		cookie, err := incomingRequest.Cookie(node.SessionCookieName)
		//debug("looking for", node.SessionCookieName, "for", node.Url, "in", incomingRequest.Cookies(), err, cookie)
		if err == nil && cookie != nil && len(cookie.Value) > 0 {
			sessionId = cookie.Value
			availableNodes = append(availableNodes, node)
		}
	}
	if len(availableNodes) == 0 {
		// No Session ID yet, pick whatever you want
		debug("resolve node: serving random node")
		availableNodes = p.Nodes
	} else {
		debug("resolve node: found a session group")
	}
	return p.balance(availableNodes), sessionId
}

func (p *Proxy) balance(nodes []*Node) *Node {
	debug("balancing nodes", len(nodes))
	if len(nodes) > 0 {
		minLoadNode := nodes[0]
		for _, node := range nodes[1:] {
			if Debug {
				debug("	node", node.Id, node.Load())
			}
			if node.Load() < minLoadNode.Load() {
				minLoadNode = node
			}
		}
		if Debug {
			debug("	min load is on", minLoadNode.Id, minLoadNode.Load())
		}
		return minLoadNode
	}
	return nil
}
