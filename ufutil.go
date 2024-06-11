package ufutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

var (
	client = http.Client{
		Timeout: 15 * time.Second, //Re-utilizamos o cliente HTTP, já que ele é thread-safe e deve ser re-usado.
	}
	userAgent = fmt.Sprintf("Mozilla/5.0 (%v; %v); ufutil-lib/v0.0.1 (%v; %v); +(https://github.com/data-ru/ufutil)", runtime.GOOS, runtime.GOARCH, runtime.Compiler, runtime.Version())
	baseUrl   = "https://sso.ufu.br"
)

type IdUfu struct {
	Cpf           string   `json:"cpf"`
	Nome          string   `json:"nome"`
	Chave         string   `json:"chave"`
	Email         string   `json:"email"`
	ExpiraEm      int      `json:"expira_em"`
	IDPessoa      int      `json:"id_pessoa"`
	EmitidoEm     int      `json:"emitido_em"`
	AccessTokenID string   `json:"access_token_id"`
	Roles         []string `json:"roles"`
	Perfis        string   `json:"perfis"`
}

type Usuario struct {
	Usuario string `json:"uid"`
	Senha   string `json:"senha"`
}

func Login(infoUsuario Usuario) (*IdUfu, error) {
	if infoUsuario.Usuario == "" || infoUsuario.Senha == "" {
		return nil, errors.New("usuario ou senha estão vazios")
	}
	userAndPass, _ := json.Marshal(infoUsuario)
	bufferJsonReq := bytes.NewReader(userAndPass)

	requestCreateLogin, err := http.NewRequest(http.MethodPost, baseUrl+"/autenticar", bufferJsonReq)
	if err != nil {
		return nil, err
	}
	requestCreateLogin.Header.Add("Content-Type", "application/json")
	requestCreateLogin.Header.Add("User-Agent", userAgent)

	responseCreateLogin, err := client.Do(requestCreateLogin)
	if err != nil {
		return nil, err
	}
	defer responseCreateLogin.Body.Close()

	body, err := io.ReadAll(responseCreateLogin.Body)
	if err != nil {
		return nil, err
	}
	responseCreateLoginBody := string(body)

	if responseCreateLogin.StatusCode != 201 {
		return nil, fmt.Errorf("algo deu errado, status http: %v, mensagem do servidor: %v", responseCreateLogin.StatusCode, responseCreateLoginBody)
	}

	getUserPath := strings.ReplaceAll(responseCreateLoginBody, "cliente-login", "usuario") //Substitui /cliente-login?t=XXXXXXXX por /usuario?t=XXXXXXXXXX

	requestGetUser, err := http.NewRequest(http.MethodGet, baseUrl+getUserPath, nil)
	if err != nil {
		return nil, err
	}
	cookiesCreate := responseCreateLogin.Cookies()
	for _, v := range cookiesCreate {
		requestGetUser.AddCookie(v) //Adiciona os cookies da request anterior
	}
	requestGetUser.Header.Add("User-Agent", userAgent)

	responseGetUser, err := client.Do(requestGetUser)
	if err != nil {
		return nil, err
	}
	defer responseGetUser.Body.Close()

	body, err = io.ReadAll(responseGetUser.Body)
	if err != nil {
		return nil, err
	}
	//responseGetUserBody := string(body)
	if responseGetUser.StatusCode != 200 {
		return nil, fmt.Errorf("algo deu errado ao obter as informações do usuario, status http %v, (%v)", responseGetUser.Status, err)
	}

	var informaçõesUsuario IdUfu
	err = json.Unmarshal(body, &informaçõesUsuario)
	if err != nil {
		return nil, err
	}

	return &informaçõesUsuario, nil
}
