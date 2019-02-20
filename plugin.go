package main

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/jackspirou/syscerts"
	"github.com/pkg/errors"
)

type (
	Config struct {
		Source        string
		Destination   string
		Authorization string
		Username      string
		Password      string
		SkipVerify    bool
		MD5           string
		SHA256        string
	}

	Plugin struct {
		Config Config
	}
)

func (p Plugin) Exec() error {
	destination := p.Config.Destination

	u, err := url.Parse(p.Config.Source)
	if err != nil {
		return errors.Wrap(err, "parsing source failed")
	}

	switch {
	case destination == "":
		destination = path.Base(u.Path)
	case destination[len(destination)-1] == filepath.Separator:
		destination = destination + path.Base(u.Path)
		fallthrough
	default:
		err = os.MkdirAll(filepath.Dir(destination), os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "parsing destination failed")
		}
	}

	log.Printf("downloading to %s", destination)

	client := &http.Client{
		Timeout: time.Minute * 5,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				RootCAs:            syscerts.SystemRootsPool(),
				InsecureSkipVerify: p.Config.SkipVerify,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if p.Config.Username != "" && p.Config.Password != "" {
				req.SetBasicAuth(p.Config.Username, p.Config.Password)
			}
			if p.Config.Authorization != "" {
				req.Header.Add("Authorization", p.Config.Authorization)
			}
			return nil
		},
	}

	req, err := http.NewRequest(
		"GET",
		p.Config.Source,
		nil,
	)

	if err != nil {
		return errors.Wrap(err, "initializing request failed")
	}

	if p.Config.Username != "" && p.Config.Password != "" {
		req.SetBasicAuth(p.Config.Username, p.Config.Password)
	}

	if p.Config.Authorization != "" {
		req.Header.Add("Authorization", p.Config.Authorization)
	}

	resp, err := client.Do(req)

	if err != nil {
		return errors.Wrap(err, "executing request failed")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("request failed, status %s", http.StatusText(resp.StatusCode))
	}

	target, err := os.Create(destination)

	if err != nil {
		return errors.Wrap(err, "creating destination failed")
	}

	defer target.Close()

	_, err = io.Copy(target, resp.Body)

	if err != nil {
		return errors.Wrap(err, "copying destination failed")
	}

	if p.Config.MD5 != "" {
		h := md5.New()
		target.Seek(0, 0)

		if _, err := io.Copy(h, target); err != nil {
			defer os.Remove(target.Name())
			return errors.Wrap(err, "failed to compare checksum")
		}

		check := fmt.Sprintf("%x", h.Sum(nil))

		if p.Config.MD5 != check {
			defer os.Remove(target.Name())
			return fmt.Errorf("checksum doesn't match, got %s and expected %s", check, p.Config.MD5)
		}
	}

	if p.Config.SHA256 != "" {
		h := sha256.New()
		target.Seek(0, 0)

		if _, err := io.Copy(h, target); err != nil {
			defer os.Remove(target.Name())
			return errors.Wrap(err, "failed to compare checksum")
		}

		check := fmt.Sprintf("%x", h.Sum(nil))

		if p.Config.SHA256 != check {
			defer os.Remove(target.Name())
			return fmt.Errorf("checksum doesn't match, got %s and expected %s", check, p.Config.SHA256)
		}
	}

	return nil
}
