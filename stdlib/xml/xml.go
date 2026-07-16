package xml

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/natanfeitosa/portuscript/ptst"
)

type XMLNode struct {
	XMLName xml.Name
	Content string    `xml:",chardata"`
	Nodes   []XMLNode `xml:",any"`
}

func parseXMLNode(n XMLNode) ptst.Objeto {
	if len(n.Nodes) == 0 {
		val := strings.TrimSpace(n.Content)
		return ptst.Texto(val)
	}
	mapa := ptst.NewMapaVazio()
	for _, child := range n.Nodes {
		mapa.M__define_item__(ptst.Texto(child.XMLName.Local), parseXMLNode(child))
	}
	return mapa
}

func met_xml_analisar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("analisar", false, args, 1, 1); err != nil {
		return nil, err
	}

	texto, err := ptst.NewTexto(args[0])
	if err != nil {
		return nil, err
	}

	var root XMLNode
	err = xml.Unmarshal([]byte(texto.(ptst.Texto)), &root)
	if err != nil {
		return nil, ptst.NewErroF(ptst.ValorErro, "Erro ao analisar XML: %v", err)
	}

	mapa := ptst.NewMapaVazio()
	mapa.M__define_item__(ptst.Texto(root.XMLName.Local), parseXMLNode(root))
	return mapa, nil
}

func met_xml_serializar(inst ptst.Objeto, args ptst.Tupla) (ptst.Objeto, error) {
	if err := ptst.VerificaNumeroArgumentos("serializar", false, args, 1, 2); err != nil {
		return nil, err
	}

	mapa, ok := args[0].(ptst.Mapa)
	if !ok {
		return nil, ptst.NewErroF(ptst.TipagemErro, "serializar esperava um Mapa como primeiro argumento")
	}

	rootTag := "raiz"
	if len(args) == 2 {
		tagText, err := ptst.NewTexto(args[1])
		if err != nil {
			return nil, err
		}
		rootTag = string(tagText.(ptst.Texto))
	}

	var builder strings.Builder
	builder.WriteByte('<')
	builder.WriteString(rootTag)
	builder.WriteByte('>')
	serializeMap(mapa, &builder)
	builder.WriteString("</")
	builder.WriteString(rootTag)
	builder.WriteByte('>')

	return ptst.Texto(builder.String()), nil
}

func serializeMap(mapa ptst.Mapa, builder *strings.Builder) {
	for k, v := range mapa {
		builder.WriteByte('<')
		builder.WriteString(k)
		builder.WriteByte('>')
		switch child := v.(type) {
		case ptst.Mapa:
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
	ptst.RegistraModuloImpl(&ptst.ModuloImpl{
		Info: ptst.ModuloInfo{
			Nome:    "xml",
			Arquivo: "stdlib/xml",
		},
		Metodos: []*ptst.Metodo{
			ptst.NewMetodoOuPanic("analisar", met_xml_analisar, ""),
			ptst.NewMetodoOuPanic("serializar", met_xml_serializar, ""),
		},
	})
}
