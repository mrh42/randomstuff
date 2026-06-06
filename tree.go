package main

import (
	"bytes"
	"github.com/iand/gedcom"
	"io/ioutil"
	"fmt"
	"os"
	"encoding/json"
)

type Data struct {
	g *gedcom.Gedcom
	rootid string
	individuals map[string]*gedcom.IndividualRecord
	
}

func NewData(filename string) (data *Data) {
	data = &Data{}

	raw, _ := ioutil.ReadFile(filename)
	d := gedcom.NewDecoder(bytes.NewReader(raw))
	data.g, _ = d.Decode()

	data.individuals = make(map[string]*gedcom.IndividualRecord)
	data.rootid = data.g.Individual[0].Xref

	for _, rec := range data.g.Individual {
		data.individuals[rec.Xref] = rec
	}
	return 
}
func (d *Data) Name(id string) (n string) {
	i := d.individuals[id]
	if i == nil { return }

	n = i.Name[0].Name
	return
}

type InfoS struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Birth       string `json:"birth"`
	Birthplace  string `json:"birthplace"`
	Death       string `json:"death"`
	Deathplace  string `json:"deathplace"`
	Father      string `json:"father"`
	Mother      string `json:"mother"`
	Children    []string `json:"children"`
	
}
func (d *Data) Info(id string) (j string) {

	info := &InfoS{ID:id}
	info.Name = d.Name(id)
	info.Birth, info.Birthplace = d.Event(id, "BIRT")
	info.Death, info.Deathplace = d.Event(id, "DEAT")
	info.Children = d.Children(id)
	info.Mother = d.Mother(id)
	info.Father = d.Father(id)
	jd, _ := json.Marshal(info)
	j = string(jd)
	return
}

func (d *Data) Event(id, tag string) (date, place string) {
	i := d.individuals[id]
	if i == nil { return }

	ev := i.Event
	for _, e := range ev {
		if e.Tag == tag {
			date = e.Date
			place = e.Place.Name
			return
		}
	}
	return
}

func (d *Data) Children(id string) (cids []string) {
	cids = make([]string, 0, 10)

	i := d.individuals[id]
	if i == nil { return }
	fs := i.Family

	for _, fl := range fs {
		f := fl.Family
		for _, c := range f.Child {
			cids = append(cids, c.Xref)
		}
	}
	return
}

func (d *Data) Father(id string) (fid string) {
	i := d.individuals[id]
	if i == nil { return }
	pf := i.Parents
	if len(pf) > 0 {
		f := pf[0].Family
		if f.Husband != nil {
			fid = f.Husband.Xref
		}
	}
	return
}
func (d *Data) Mother(id string) (wid string) {
	i := d.individuals[id]
	if i == nil { return }
	pf := i.Parents
	if len(pf) > 0 {
		f := pf[0].Family
		if f.Wife != nil {
			wid = f.Wife.Xref
		}
	}
	return
}
func main() {
	
	d := NewData("mrh-tree.ged")
	ids := make([]string, 0, 10)
	if len(os.Args) > 1 {
		for _, x := range os.Args[1:] {
			ids = append(ids, x)
		}
	}
	if len(ids) == 0 {
		ids = append(ids, d.rootid)
	}
	for _, id := range(ids) {
		j := d.Info(id)
		fmt.Println(j)
	}
}
