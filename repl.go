package main

import (
	"encoding/base64"
	"io"
	"regexp"
	"time"
	//"strconv"
	"strings"

	"github.com/chzyer/readline"
	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core/chain"
	"github.com/u-speak/core/node"
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func formatHash(hash [32]byte) string {
	// if Config.Logger.PrintEmoji {
	// 	return util.CompactEmoji(hash)
	// }
	return base64.URLEncoding.EncodeToString(hash[:])
}

func startRepl(n *node.Node) {
	ci := func(s string) *readline.PrefixCompleter {
		return readline.PcItem(s)
	}
	var completer = readline.NewPrefixCompleter(
		readline.PcItem("post",
			readline.PcItem("get"),
			readline.PcItem("add"),
		),
		readline.PcItem("mine"),
		readline.PcItem("chain",
			readline.PcItem("print", ci("post"), ci("key"), ci("image")),
			readline.PcItem("validate", ci("post"), ci("key"), ci("image")),
		),
		readline.PcItem("node",
			readline.PcItem("connect"),
			readline.PcItem("status"),
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
		case strings.HasPrefix(line, "post get "):
			lc := strings.Split(line, " ")
			h, err := base64.URLEncoding.DecodeString(lc[2])
			if err != nil {
				log.Error(err)
				break
			}
			var hash [32]byte
			copy(hash[:], h)
			c := getChain(n, lc[2])
			block := c.Get(hash)
			if block == nil {
				log.Info("No block found")
				break
			}
			log.WithFields(log.Fields{
				"hash": formatHash(block.Hash()),
				"prev": formatHash(block.PrevHash),
				"date": block.Date.String(),
			}).Debug(block.Content)
		case strings.HasPrefix(line, "post add "):
			lc := strings.Split(line, " ")
			content := lc[2]
			pc := n.PostChain
			b := chain.Block{Date: time.Now(), Type: "post", PrevHash: pc.LastHash(), Content: content}
			h, err := n.PostChain.Add(b)
			if err != nil {
				log.Error(err)
			} else {
				log.WithField("hash", base64.URLEncoding.EncodeToString(h[:])).Info("Block added")
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
			log.Infof("Staus for node %s", s.Address)
			log.Infof("Total Blocks: %d", s.Length)
			log.WithFields(log.Fields{
				"Length":   s.Chains.Post.Length,
				"Valid":    s.Chains.Post.Valid,
				"LastHash": s.Chains.Post.LastHash,
			}).Info("Post Chain")
			log.WithFields(log.Fields{
				"Length":   s.Chains.Image.Length,
				"Valid":    s.Chains.Image.Valid,
				"LastHash": s.Chains.Image.LastHash,
			}).Info("Image Chain")
			log.WithFields(log.Fields{
				"Length":   s.Chains.Key.Length,
				"Valid":    s.Chains.Key.Valid,
				"LastHash": s.Chains.Key.LastHash,
			}).Info("Key Chain")
			log.Info("End Status")
		case simpleMatch("chain print (post|image|key)", line):
			c := getChain(n, strings.Split(line, " ")[2])
			dump, err := c.DumpChain()
			if err != nil {
				log.Error(err)
				break
			}
			for _, b := range dump {
				log.WithFields(log.Fields{
					"hash": formatHash(b.Hash()),
					"prev": formatHash(b.PrevHash),
				}).Debug(b.Content)
			}
		case simpleMatch("chain validate (post|image|key)", line):
			c := getChain(n, strings.Split(line, " ")[2])
			if c.Valid() {
				log.Info("Chain is valid")
			} else {
				log.Error("Chain is invalid")
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

func getChain(n *node.Node, t string) *chain.Chain {
	switch t {
	case "post":
		return n.PostChain
	case "image":
		return n.ImageChain
	case "key":
		return n.KeyChain
	default:
		return n.PostChain
	}
}
