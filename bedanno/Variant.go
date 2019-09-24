package main

import (
	"encoding/json"
	"github.com/brentp/irelate/interfaces"
	"github.com/brentp/vcfgo"
	simple_util "github.com/liserjrqlxue/simple-util"
	"strings"
)

type Variant struct {
	Chromosome string
	Pos        uint64
	Id_        string
	Reference  string
	Alternate  []string
	Quality    float32
	Filter     string
	Info_      interfaces.Info
	Format     []string
	LineNumber int64
}

// String gives a string representation of a variant
func (v *Variant) String() string {
	return v.Info().String()
}
func (v *Variant) Chrom() string {
	return v.Chromosome
}

// Start returns the 0-based start
func (v *Variant) Start() uint32 {
	return uint32(v.Pos - 1)
}

func (v *Variant) Ref() string {
	return v.Reference
}

func (v *Variant) Alt() []string {
	return v.Alternate
}

// End returns the 0-based start + the length of the reference allele.
func (v *Variant) End() uint32 {
	return uint32(v.Pos-1) + uint32(len(v.Ref()))
}

func (v *Variant) Info() interfaces.Info {
	return v.Info_
}

func (v *Variant) Id() string {
	return v.Id_
}

// CIPos reports the Left and Right end of an SV using the CIPOS tag. It is in
// bed format so the end is +1'ed. E.g. If there is not CIPOS, the return value
// is v.Start(), v.Start() + 1
func (v *Variant) CIPos() (uint32, uint32, bool) {
	s := v.Start()
	return s, s + 1, false
}

// CIEnd reports the Left and Right end of an SV using the CIEND tag. It is in
// bed format so the end is +1'ed. E.g. If there is no CIEND, the return value
// is v.End() - 1, v.End()
func (v *Variant) CIEnd() (uint32, uint32, bool) {
	e := v.End()
	return e - 1, e, false
}

type InfoByte struct {
	Info   map[string]string
	header []string
}

func NewInfoByte(info []string, h []string) *InfoByte {
	var item = make(map[string]string)
	for i, key := range h {
		if i < len(info) {
			item[key] = string(info[i])
		}
	}
	return &InfoByte{Info: item, header: h}
}

func (i InfoByte) Get(key string) (interface{}, error) {
	v := i.Info[key]
	return v, nil
}

func (i InfoByte) Set(key string, value interface{}) error {
	i.Info[key] = vcfgo.ItoS(key, value)
	return nil
}
func (i InfoByte) Bytes() []byte {
	jsonBytes, err := json.Marshal(i.Info)
	simple_util.CheckErr(err)
	return jsonBytes
}

func (i InfoByte) String() string {
	var values []string
	for _, key := range i.header {
		values = append(values, i.Info[key])
	}
	return strings.Join(values, "\t")
}

func (i InfoByte) Delete(key string) {
	i.Info[key] = ""
	return
}
func (i InfoByte) Keys() []string {
	var title []string
	for key := range i.Info {
		title = append(title, key)
	}
	return title
}
