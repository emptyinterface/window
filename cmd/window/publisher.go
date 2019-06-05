package main

import "sync"

type (
	Publisher struct {
		nodes map[string]map[chan string]struct{}
		me    sync.Mutex
		f     func(path string) string
	}
)

func NewPublisher(f func(path string) string) *Publisher {
	return &Publisher{
		nodes: map[string]map[chan string]struct{}{},
		me:    sync.Mutex{},
		f:     f,
	}
}

func (p *Publisher) Subscribe(path string, node chan string) {
	p.me.Lock()
	defer p.me.Unlock()
	select {
	case node <- p.f(path):
	default:
	}
	if _, exists := p.nodes[path]; !exists {
		p.nodes[path] = map[chan string]struct{}{}
	}
	p.nodes[path][node] = struct{}{}
}

func (p *Publisher) Unsubscribe(path string, node chan string) {
	p.me.Lock()
	defer p.me.Unlock()
	delete(p.nodes[path], node)
	if len(p.nodes[path]) == 0 {
		delete(p.nodes, path)
	}
}

func (p *Publisher) Publish() {
	p.me.Lock()
	defer p.me.Unlock()
	for path, nodes := range p.nodes {
		data := p.f(path)
		for node, _ := range nodes {
			select {
			case node <- data:
			default:
			}
		}
	}
}
