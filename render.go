package main

import (
	"bytes"
	"text/template"

	"github.com/kenshaw/snaker"
	"github.com/yangsai7/go-dao-gen-code/tplbin"
)

type Indexes map[string][]string // index name -> columns

type RenderData struct {
	Pkg                  string
	Table                string
	TableLowerCamelIdent string
	TableUpperCamelIdent string
	Primary              string
	Attrs                []*AttrEntity
	UniqueIndexes        Indexes
	ShadowTables         map[string]string
}

func RenderTable(name string, data *RenderData) (content []byte, err error) {
	tpl, err := tplbin.Asset("table.tpl")
	if err != nil {
		return content, err
	}
	t, err := template.New(name).Funcs(template.FuncMap{
		"ToUpperCamel": snaker.SnakeToCamelIdentifier,
	}).Parse(string(tpl))
	if err != nil {
		return content, err
	}
	buf := bytes.NewBuffer(content)
	err = t.Execute(buf, data)
	return buf.Bytes(), err
}

func RenderTableConds(name string, data *RenderData) (content []byte, err error) {
	tpl, err := tplbin.Asset("conds.tpl")
	if err != nil {
		return content, err
	}
	t, err := template.New(name).Parse(string(tpl))
	if err != nil {
		return content, err
	}
	buf := bytes.NewBuffer(content)
	err = t.Execute(buf, data)
	return buf.Bytes(), err
}

func RenderInitDao(data *RenderData) (content []byte, err error) {
	tpl, err := tplbin.Asset("dao.tpl")
	if err != nil {
		return content, err
	}
	t, err := template.New("dao").Parse(string(tpl))
	if err != nil {
		return content, err
	}
	buf := bytes.NewBuffer(content)
	err = t.Execute(buf, data)
	return buf.Bytes(), err
}
