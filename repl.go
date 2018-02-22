package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/u-speak/core/api"
	"github.com/u-speak/core/node"
	"github.com/u-speak/core/post"
	"github.com/u-speak/core/tangle"
	"github.com/u-speak/core/tangle/hash"
	"github.com/u-speak/core/tangle/site"

	"github.com/chzyer/readline"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func startRepl(n *node.Node) {
	var completer = readline.NewPrefixCompleter(
		readline.PcItem("site",
			readline.PcItem("get"),
			readline.PcItem("add"),
		),
		readline.PcItem("tangle",
			readline.PcItem("print"),
			readline.PcItem("status"),
		),
		readline.PcItem("node",
			readline.PcItem("connect"),
			readline.PcItem("status"),
			readline.PcItem("merge"),
		),
	)
	l, err := readline.NewEx(&readline.Config{
		Prompt:          n.ListenInterface + " \033[31m»\033[0m ",
		HistoryFile:     "/tmp/uspeakd-repl.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "site get "):
			lc := strings.Split(line, " ")
			h, err := base64.URLEncoding.DecodeString(lc[2])
			if err != nil {
				log.Error(err)
				break
			}
			s := n.Tangle.Get(hash.FromSlice(h))
			vals := []hash.Hash{}
			for _, v := range s.Site.Validates {
				vals = append(vals, v.Hash())
			}
			log.WithFields(log.Fields{
				"hash":      s.Site.Hash(),
				"validates": vals,
				"weight":    n.Tangle.Weight(s.Site),
				"type":      s.Site.Type,
			}).Debug(s.Site.Content)
		case strings.HasPrefix(line, "data get "):
			lc := strings.Split(line, " ")
			h, err := base64.URLEncoding.DecodeString(lc[2])
			if err != nil {
				log.Error(err)
				break
			}
			s := n.Tangle.Get(hash.FromSlice(h))
			switch s.Site.Type {
			case "post":
				p := s.Data.(*post.Post)
				log.WithFields(log.Fields{
					"date":  p.Timestamp,
					"valid": p.Verify() == nil,
					"keyid": p.Pubkey.KeyIdShortString(),
				}).Info(p.Content)
			}
		case strings.HasPrefix(line, "site add "):
			cnt := line[9:]
			recs := n.Tangle.RecommendTips()
			for _, r := range recs {
				log.Infof("Recommended: %s", r.Hash())
			}
			post := genpost(cnt)
			h, err := post.Hash()
			if err != nil {
				log.Error(err)
				break
			}
			s := &site.Site{
				Validates: recs,
				Content:   h,
				Type:      "post",
			}
			s.Mine(1)
			log.WithFields(log.Fields{"nonce": s.Nonce, "weight": s.Hash().Weight()}).Infof("Finished Mining: %s", s.Hash())

			n.Submit(&tangle.Object{Site: s, Data: post})
			if err != nil {
				log.Error(err)
			}
		case strings.HasPrefix(line, "site gen "):
			cnt := line[9:]
			recs := n.Tangle.RecommendTips()
			post := genpost(cnt)
			h, err := post.Hash()
			if err != nil {
				log.Error(err)
				break
			}
			s := &site.Site{
				Validates: recs,
				Content:   h,
				Type:      "post",
			}
			s.Mine(1)
			j := api.JSONize(&tangle.Object{Site: s, Data: post})
			b, err := json.Marshal(j)
			if err != nil {
				log.Error(err)
				break
			}
			fmt.Println(string(b))
		case strings.HasPrefix(line, "node merge "):
			remote := strings.Split(line, " ")[2]
			err := n.Merge(remote)
			if err != nil {
				log.Error(err)
				break
			}
		case strings.HasPrefix(line, "node connect"):
			remote := strings.Split(line, " ")[2]
			err := n.Connect(remote)
			if err != nil {
				log.Error(err)
			} else {
				log.Info("Successfully connected")
			}
		case line == "node status":
			s := n.Status()
			printInfo(&s)
		case line == "tangle status":
			log.Info("--- Tangle Status ---")
			log.Info("Tips:")
			for _, t := range n.Tangle.Tips() {
				log.Info("  ", t.Hash())
			}
			log.Info("--- End Status ---")
		case strings.HasPrefix(line, "node status "):
			remote := strings.Split(line, " ")[2]
			i, err := n.RemoteStatus(remote)
			if err != nil {
				log.Error(err)
				break
			}
			printInfo(i)
		case line == "tangle print":
			c := 0
			for _, h := range n.Tangle.Hashes() {
				if c == 10 {
					break
				}
				log.Info(h)
				c++
			}
		case line == "exit":
			return
		case line == "":
			break
		default:
			log.Warnf("Command `%s' not found", line)
			log.Warn("Please check if you specified the correct number of arguments")
		}
	}
}

func simpleMatch(pattern, s string) bool {
	m, err := regexp.MatchString(pattern, s)
	if err != nil {
		log.Error(err)
	}
	return m
}

func genpost(c string) *post.Post {
	content := c
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	privkey := packet.NewRSAPrivateKey(time.Now(), key)
	buff := bytes.NewBuffer(nil)
	e := &openpgp.Entity{
		PrivateKey: privkey,
		PrimaryKey: &privkey.PublicKey,
	}
	_ = openpgp.ArmoredDetachSignText(buff, e, strings.NewReader(content), nil)
	block, _ := armor.Decode(buff)
	reader := packet.NewReader(block.Body)
	pkt, _ := reader.Next()
	sig, _ := pkt.(*packet.Signature)
	p := &post.Post{Content: content, Pubkey: &privkey.PublicKey, Signature: sig, Timestamp: time.Now().Unix()}
	_ = p.JSON()
	return p
}

func printInfo(s *node.Status) {
	log.Infof("--- Status for node %s ---", s.Address)
	log.Infof("Total Sites: %d", s.Length)
	log.Infof("Remote Connections: %d", s.Connections)
	log.Infof("--- BEGIN DIFF ---")
	for _, h := range s.HashDiff.Additions {
		log.Infof("+ %s", h.String())
	}
	for _, h := range s.HashDiff.Deletions {
		log.Infof("- %s", h.String())
	}
	log.Info("--- End Status ---")
}
