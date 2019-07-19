package acm

import (
	"log"
	"testing"
)

func TestAcm_New(t *testing.T) {
	a := New()
	log.Println(a.m)
	log.Println(a.m["public"])
}

func TestAcm_selectIP(t *testing.T) {
	a := New()
	p := Params{
		ServerDomainAddr: HZ,
	}
	ip := a.selectIP(p)
	t.Log(ip)
}
