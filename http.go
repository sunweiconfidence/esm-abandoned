/*
Copyright 2016 Medcl (m AT medcl.net)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"net/http"
	"github.com/parnurzeal/gorequest"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"io"
	"errors"
	"bytes"
	"net/url"
	"crypto/tls"
)

func Get(url string,auth *Auth,proxy string) (*http.Response, string, []error) {
	request := gorequest.New()

	tr := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	request.Transport=tr


	if(auth!=nil){
		request.SetBasicAuth(auth.User,auth.Pass)
	}

	request.Header["Content-Type"] = []string { "application/json" }

	if(len(proxy)>0){
		request.Proxy(proxy)
	}

	resp, body, errs := request.Get(url).End()
	return resp, body, errs

}

func Post(url string,auth *Auth, body string,proxy string)(*http.Response, string, []error)  {
	request := gorequest.New()
	tr := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	request.Transport=tr

	if(auth!=nil){
		request.SetBasicAuth(auth.User,auth.Pass)
	}

	request.Header["Content-Type"] = []string { "application/json" }
	
	if(len(proxy)>0){
		request.Proxy(proxy)
	}

	request.Post(url)

	if(len(body)>0) {
		request.Send(body)
	}

	return request.End()
}

func newDeleteRequest(client *http.Client,method, urlStr string) (*http.Request, error) {
	if method == "" {
		// We document that "" means "GET" for Request.Method, and people have
		// relied on that from NewRequest, so keep that working.
		// We still enforce validMethod for non-empty methods.
		method = "GET"
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method:     method,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       u.Host,
	}
	return req, nil
}

func Request(method string,r string,auth *Auth,body *bytes.Buffer,proxy string)(string,error)  {

	//TODO use global client
	var client *http.Client
	client = &http.Client{}
	if(len(proxy)>0){
		proxyURL, err := url.Parse(proxy)
		if(err!=nil){
			log.Error(err)
		}else{
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
				DisableKeepAlives: true,
			}
			client = &http.Client{Transport: transport}
		}
	}

	tr := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
	},
	}

	client.Transport=tr

	var reqest *http.Request
	if(body!=nil){
		reqest, _ =http.NewRequest(method,r,body)
	}else{
		reqest, _ = newDeleteRequest(client,method,r)
	}
	if(auth!=nil){
		reqest.SetBasicAuth(auth.User,auth.Pass)
	}

	reqest.Header.Set("Content-Type", "application/json")


	resp,errs := client.Do(reqest)
	if errs != nil {
		log.Error(errs)
		return "",errs
	}

	if resp!=nil&& resp.Body!=nil{
		//io.Copy(ioutil.Discard, resp.Body)
		defer resp.Body.Close()
	}

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return "",errors.New("server error: "+string(b))
	}

	respBody,err:=ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error(err)
		return string(respBody),err
	}

	log.Trace(r,string(respBody))

	if err != nil {
		return string(respBody),err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	return string(respBody),nil
}
