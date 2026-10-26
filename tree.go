package main

import (
	"bytes"
	"github.com/iand/gedcom"
	"io/ioutil"
	"fmt"
	"strings"
	"encoding/json"
	"github.com/lithammer/fuzzysearch/fuzzy"
)


type Data struct {
	g           *gedcom.Gedcom
	ixrefs      map[string]int
	names       []string
	births      []string
	deaths      []string
	raw         []byte
}

func NewData(filename string) (data *Data) {
	data = &Data{}

	data.raw, _ = ioutil.ReadFile(filename)
	d := gedcom.NewDecoder(bytes.NewReader(data.raw))
	data.g, _ = d.Decode()

	data.ixrefs = make(map[string]int)
	// allows searching names
	data.names = make([]string, len(data.g.Individual))

	for i, rec := range data.g.Individual {
		data.ixrefs[rec.Xref] = i
		//fmt.Printf("%d: %s\n", i, rec.Xref)
		data.names[i] = rec.Name[0].Name
	}

	// allow searching birth and death places
	data.births = make([]string, len(data.names))
	data.deaths = make([]string, len(data.names))
	for i := range data.names {
		_, data.births[i] = data.Event(i, "BIRT")
		_, data.deaths[i] = data.Event(i, "DEAT")
	}
	fmt.Printf("Tree contains %d individuals\n", len(data.ixrefs))
	return
}

// extreme hack for now...
func (d *Data) RawData(xref1 string) (r string) {
	startm := fmt.Sprintf("0 @%s@ INDI", xref1)

	raws := string(d.raw)
	i := strings.Index(raws, startm)

	raws = raws[i:]
	i = strings.Index(raws, "\n0 @")
	r = raws[:i]

	return 
}

func (d *Data) idx(xref string) (id int) {
	id, ok := d.ixrefs[xref]
	if !ok {
		id = -1
	}
	return
}
func (d *Data) ind(id int) (*gedcom.IndividualRecord) {
	if id < 0 {return nil}
	return d.g.Individual[id]
}

func (d *Data) Name(id int) (n string) {
	i := d.ind(id)
	if i == nil { return }

	n = i.Name[0].Name
	return
}
func (d *Data) Sex(id int) (s string) {
	i := d.ind(id)
	if i == nil { return }

	s = i.Sex
	return
}

type Spouse struct {
	ID          int    `json:"id"`
	Married     string `json:"married"`
	Divorced    string `json:"divorced"`
}
type InfoDetail struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Sex         string `json:"sex"`
	Birth       string `json:"birth"`
	Birthplace  string `json:"birthplace"`
	Death       string `json:"death"`
	Deathplace  string `json:"deathplace"`
	Father      int    `json:"father"`
	Mother      int    `json:"mother"`
	Children    []int  `json:"children"`
	Spouses     []Spouse  `json:"spouses"`
	
}
func (d *Data) Info(id int) (j string) {

	info := &InfoDetail{ID:id}
	info.Name = d.Name(id)
	info.Sex = d.Sex(id)
	info.Birth, info.Birthplace = d.Event(id, "BIRT")
	info.Death, info.Deathplace = d.Event(id, "DEAT")
	info.Children = d.Children(id)
	info.Spouses = d.Spouses(id)
	info.Mother = d.Mother(id)
	info.Father = d.Father(id)
	jd, _ := json.Marshal(info)
	j = string(jd) + "\n"
	return
}


type InfoG struct {
	ID          int    `json:"id"`
	GED         string  `json:"ged"`
}

func (d *Data) GEDCOMInfo(id int) (j string) {
	ged := &InfoG{ID:id}

	i := d.ind(id)
	if i != nil {
		ged.GED = d.RawData(i.Xref)
	}
	jd, _ := json.Marshal(ged)
	j = string(jd) + "\n"
	return
}


func (d *Data) Event(id int, tag string) (date, place string) {
	i := d.ind(id)
	if i == nil { return }

	ev := i.Event
	for _, e := range ev {
		//fmt.Printf("event: id: %d tag: %s type: %s date: %s place: %s\n", id, e.Tag, e.Type, e.Date, e.Place.Name)
		if e.Tag == tag {
			date = e.Date
			place = e.Place.Name
			return
		}
	}
	return
}

func (d *Data) Spouses(id int) (spouses []Spouse) {
	spouses = make([]Spouse, 0, 10)

	i := d.ind(id)
	if i == nil { return }
	fs := i.Family

	for _, fl := range fs {
		f := fl.Family
		if f == nil {continue}

		var spouse Spouse
		for _, fe := range f.Event {
			//fmt.Printf("id: %d, tag: %s date: %s place: %s\n", id, fe.Tag, fe.Date, fe.Place)
			if fe.Tag == "MARR" {
				spouse.Married = fmt.Sprintf("%s %s", fe.Date, fe.Place.Name)
			}
			if fe.Tag == "DIV" {
				spouse.Divorced = fmt.Sprintf("%s %s", fe.Date, fe.Place.Name)
			}
		}

		if f.Wife != nil {
			wifeid := d.idx(f.Wife.Xref)
			if wifeid >= 0 && wifeid != id {
				spouse.ID = wifeid
				spouses = append(spouses, spouse)
			}
		}
		if f.Husband != nil {
			husbandid := d.idx(f.Husband.Xref)
			if husbandid >= 0 && husbandid != id {
				spouse.ID = husbandid
				spouses = append(spouses, spouse)
			}
		}
	}
	return
}

func (d *Data) Children(id int) (cids []int) {
	cids = make([]int, 0, 10)

	i := d.ind(id)
	if i == nil { return }
	fs := i.Family

	for _, fl := range fs {
		f := fl.Family
		for _, c := range f.Child {
			cid := d.idx(c.Xref)
			//fmt.Printf("cid: %d x: %s\n", cid, c.Xref)
			cids = append(cids, cid)
		}
	}
	return
}

// only returns one father, but good enough for demo
func (d *Data) Father(id int) (fid int) {
	fid = -1
	i := d.ind(id)
	if i == nil { return }
	pf := i.Parents
	for _, p := range pf {
		f := p.Family
		if f.Husband != nil {
			fid = d.idx(f.Husband.Xref)
			if fid >= 0 {
				break
			}
		}
	}
	return
}

// only returns one mother, but good enough for demo
func (d *Data) Mother(id int) (mid int) {
	mid = -1
	i := d.ind(id)
	if i == nil { return }
	pf := i.Parents
	for _, p := range pf {
		f := p.Family
		if f.Wife != nil {
			mid = d.idx(f.Wife.Xref)
			if mid >= 0 {
				break
			}
		}
	}
	return
}

func (d *Data) Ancestors(id int) (a map[int]bool) {

	a = make(map[int]bool)
	todo := make([]int, 1, 1000)
	todo[0] = id

	for len(todo) > 0 {
		l := len(todo) - 1
		id = todo[l]
		todo = todo[:l]

		p := d.Mother(id)
		if p >= 0 {
			todo = append(todo, p)
			a[p] = true
		}
		p = d.Father(id)
		if p >= 0 {
			todo = append(todo, p)
			a[p] = true
		}
	}
	return
}

func (d *Data) Common(ida, idb int) (cl []int) {
	a := d.Ancestors(ida)
	b := d.Ancestors(idb)
	c := make(map[int]bool)
	
	for i := range a {
		if b[i] {
			c[i] = true
		}
	}
	cl = make([]int, 0, len(c))
	for i := range c {
		cl = append(cl, i)
	}
	return
}

// quick hack to let the llm search
func (d *Data) Search(field, name string) (ids map[int]bool) {

	ids = make(map[int]bool)
	//name, _ = strconv.Unquote(name)

	var target []string
	switch field {
	case "BIRTH":
		target = d.births
	case "DEATH":
		target = d.deaths
	default:
		target = d.names
	}
	matches := fuzzy.RankFind(name, target)
	for i, m := range matches {
		ids[m.OriginalIndex] = true

		// limit how many people we will add
		if i > 200 {
			break
		}
	}
	fmt.Printf("Search(%s) added %d id entries\n", name, len(ids))
	return
}

