package main

import (
	"bufio"
	"fmt"
	"github.com/brentp/irelate/interfaces"
	"github.com/brentp/irelate/parsers"
	"github.com/brentp/vcfgo"
	"io"
	"log"
	"strconv"
	"strings"
)

type Writer struct {
	io.Writer
	Header []string
}

func NewWriter(w io.Writer, h []string) (*Writer, error) {
	_, err := fmt.Fprintln(w, strings.Join(h, "\t"))
	return &Writer{w, h}, err
}

type Reader struct {
	buf        *bufio.Reader
	r          io.Reader
	verr       *vcfgo.VCFError
	LineNumber int64
	Header     []string
}

func (vr *Reader) AddInfoToHeader(id string, itype string, number string, description string) {
	vr.Header = append(vr.Header, id)
}

func (vr *Reader) Read() *Variant {
	line, err := vr.buf.ReadString('\n')
	//line,err:=vr.buf.ReadBytes('\n')
	if err != nil {
		if len(line) == 0 && err == io.EOF {
			return nil
		} else if err != io.EOF {
			vr.verr.Add(err, vr.LineNumber)
		}
	}
	vr.LineNumber++
	if line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	log.Printf("line:%+v", line)
	fields := strings.Split(line, "\t")
	variant := vr.Parse(fields)
	return variant
}

func (vr *Reader) Parse(fields []string) *Variant {
	start, err := strconv.ParseUint(fields[1], 10, 64)
	vr.verr.Add(err, vr.LineNumber)
	pos := start + 1

	v := &Variant{
		Chromosome: fields[0],
		Pos:        pos,
		Reference:  fields[3],
		Alternate:  strings.Split(string(fields[4]), ","),
	}
	v.LineNumber = vr.LineNumber
	v.Info_ = NewInfoByte(fields, vr.Header)
	return v
}

func (vr *Reader) Error() error {
	if vr.verr.IsEmpty() {
		return nil
	}
	return vr.verr
}

func Bopen(rdr io.Reader) (*Reader, error) {
	buffered := bufio.NewReaderSize(rdr, 32768*2)

	var verr = vcfgo.NewVCFError()

	var LineNumber int64
	var h []string

	for {

		LineNumber++
		line, err := buffered.ReadString('\n')
		if err != nil && err != io.EOF {
			verr.Add(err, LineNumber)
		}
		if len(line) > 1 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		if LineNumber == 1 {
			h = strings.Split(line, "\t")
			break
		}
	}
	reader := &Reader{buffered, rdr, verr, LineNumber, h}
	return reader, reader.Error()
}

type bWrapper struct {
	*Reader
}

func (v bWrapper) Next() (interfaces.Relatable, error) {
	r := v.Read()
	if r == nil {
		return nil, io.EOF
	}
	return &parsers.Variant{IVariant: r}, nil
}

func (v bWrapper) Close() error {
	if rc, ok := v.r.(io.ReadCloser); ok {
		return rc.Close()
	}
	return nil
}

func BedIterator(buf io.Reader) (interfaces.RelatableIterator, *Reader, error) {
	v, err := Bopen(buf)
	if err != nil {
		return nil, v, err
	}
	return bWrapper{v}, v, nil
}
