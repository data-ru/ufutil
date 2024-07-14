package ufutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strings"
	"time"
)

var (
	client = http.Client{
		Timeout: 15 * time.Second, //Re-utilizamos o cliente HTTP, já que ele é thread-safe e deve ser re-usado.
	}
	userAgent = fmt.Sprintf("Mozilla/5.0 (%v; %v); ufutil-lib/v1.0.0 (%v; %v); +(https://github.com/data-ru/ufutil)", runtime.GOOS, runtime.GOARCH, runtime.Compiler, runtime.Version())
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
	Matricula     string
}

type Usuario struct {
	Usuario string `json:"uid"`
	Senha   string `json:"senha"`
}

type resultadoIdentidade struct {
	IdentidadeDigital        *IdentidadeDigital `json:"identidadeDigital"`
	DocumentoArquivoTOResult any                `json:"documentoArquivoTOResult"`
	DataNascimentoString     *string            `json:"dataNascimentoString"`
}
type IdentidadeDigital struct {
	ID                int    `json:"id"`
	Matricula         string `json:"matricula"`
	Nome              string `json:"nome"`
	NomePai           string `json:"nomePai"`
	NomeMae           string `json:"nomeMae"`
	Rg                string `json:"rg"`
	OrgaoEmissor      string `json:"orgaoEmissor"`
	Cpf               string `json:"cpf"`
	Naturalidade      string `json:"naturalidade"`
	Vinculo           string `json:"vinculo"`
	DataNascimento    int64  `json:"dataNascimento"`
	CodigoBarra       string `json:"codigoBarra"`
	Informacao        string `json:"informacao"`
	SituacaoDescricao string `json:"situacaoDescricao"`
	Situacao          int    `json:"situacao"`
	Foto              string `json:"foto"`
}

/*type resLoginMobile struct {
	ResultType    string      `json:"resultType"`
	ResultCode    string      `json:"resultCode"`
	Nome          string      `json:"nome"`
	Token         string      `json:"token"`
	Perfis        []perfis    `json:"perfis"`
	PerfilAtivo   perfilAtivo `json:"perfilAtivo"`
	Email         any         `json:"email"`
	Avatar        string      `json:"avatar"`
	DataExpMillis int64       `json:"dataExpMillis"`
}
type perfis struct {
	IDPerfil       int    `json:"idPerfil"`
	NomePerfil     string `json:"nomePerfil"` //matricula
	TipoPerfil     string `json:"tipoPerfil"`
	NomeTipoPerfil string `json:"nomeTipoPerfil"`
	Selecionado    bool   `json:"selecionado"`
}
type perfilAtivo struct {
	IDPerfil       int    `json:"idPerfil"`
	NomePerfil     string `json:"nomePerfil"`
	TipoPerfil     string `json:"tipoPerfil"`
	NomeTipoPerfil string `json:"nomeTipoPerfil"`
	Selecionado    bool   `json:"selecionado"`
}*/

/*type Cardapio struct {
	Lugar  map[string]int
	Pratos ApiCardapio
}*/

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

var ErrNãoHáRefeições = errors.New("não há refeições agendadas para hoje")

func CardapioDoDiaTodosCampi() (ApiCardapio, error) {
	resp, err := requisiçãoGenerica("https://www.sistemas.ufu.br/mobile-gateway/api/cardapios/", http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decryptBody, err := decryptJSON(string(bodyResp))
	if err != nil {
		return nil, ErrNãoHáRefeições
	}

	return decryptBody, nil
}

// A url deve parecer algo como: https://www.sistemas.ufu.br/valida-ufu/#/id-digital/XXXXXXXXXXXXXX
func ValidarIdUfu(urlId string) (*IdentidadeDigital, error) {

	_, err := url.Parse(urlId) //Valida a URL
	if err != nil {
		return nil, err
	}
	id := path.Base(urlId) //Obtem o ID da URL: XXXXXXXXXXXXXX

	res, err := requisiçãoGenerica("https://www.sistemas.ufu.br/valida-gateway/id-digital/buscarDadosIdDigital?idIdentidade="+id, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	corpo, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("servidor retornou o código http %v", res.StatusCode)
	}

	var jsonId resultadoIdentidade
	err = json.Unmarshal(corpo, &jsonId)
	if err != nil {
		return nil, err
	}

	if jsonId.DataNascimentoString == nil {
		return nil, errors.New("identidade invalida")
	}

	return jsonId.IdentidadeDigital, nil
}

func requisiçãoGenerica(url, meteodo string, corpo io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(meteodo, url, corpo)
	if err != nil {
		return nil, err
	}
	//req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", userAgent)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Descriptografar(json string) (string, error) {
	return decryptJsonAsString(json)
}

func Criptografar(json string) (string, error) {
	return makeRequest(json)
}
