package xml

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/mat-dgruber/Harpia/hrp"
)

type XMLNode struct {
	XMLName xml.Name
	Content string    `xml:",chardata"`
	Nodes   []XMLNode `xml:",any"`
}

func parseXMLNode(n XMLNode) hrp.Objeto {
	if len(n.Nodes) == 0 {
		val := strings.TrimSpace(n.Content)
		return hrp.Texto(val)
	}
	mapa := hrp.NewMapaVazio()
	for _, child := range n.Nodes {
		mapa.M__define_item__(hrp.Texto(child.XMLName.Local), parseXMLNode(child))
	}
	return mapa
}

func met_xml_analisar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("analisar", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := hrp.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	var root XMLNode
	err = xml.Unmarshal([]byte(texto.(hrp.Texto)), &root)
	if err != nil {
		return nil, hrp.NewErroF(hrp.ValorErro, "Erro ao analisar XML: %v", err)
	}

	mapa := hrp.NewMapaVazio()
	mapa.M__define_item__(hrp.Texto(root.XMLName.Local), parseXMLNode(root))
	return mapa, nil
}

func met_xml_serializar(inst hrp.Objeto, args hrp.Tupla) (hrp.Objeto, error) {
	if err := hrp.VerificaNumeroArgumentos("serializar", false, args, 1, 2); err != nil {
		return nil, err
	}

	mapa, ok := args[0].(hrp.Mapa)
	if !ok {
		return nil, hrp.NewErroF(hrp.TipagemErro, "serializar esperava um Mapa como primeiro argumento")
	}

	rootTag := "raiz"
	if len(args) == 2 {
		tagText, err := hrp.NewTexto(args[1])
		if err != nil {
			return nil, err
		}
		rootTag = string(tagText.(hrp.Texto))
	}

	var builder strings.Builder
	builder.WriteByte('<')
	builder.WriteString(rootTag)
	builder.WriteByte('>')
	serializeMap(mapa, &builder)
	builder.WriteString("</")
	builder.WriteString(rootTag)
	builder.WriteByte('>')

	return hrp.Texto(builder.String()), nil
}

func serializeMap(mapa hrp.Mapa, builder *strings.Builder) {
	for k, v := range mapa {
		builder.WriteByte('<')
		builder.WriteString(k)
		builder.WriteByte('>')
		switch child := v.(type) {
		case hrp.Mapa:
			serializeMap(child, builder)
		default:
			builder.WriteString(fmt.Sprintf("%v", v))
		}
		builder.WriteString("</")
		builder.WriteString(k)
		builder.WriteByte('>')
	}
}

func init() {
	hrp.RegistraModuloImpl(&hrp.ModuloImpl{
		Info: hrp.ModuloInfo{
			Nome:    "xml",
			Arquivo: "stdlib/xml",
		},
		Metodos: []*hrp.Metodo{
			hrp.NewMetodoOuPanic("analisar", met_xml_analisar, ""),
			hrp.NewMetodoOuPanic("serializar", met_xml_serializar, ""),
		},
	})
}
