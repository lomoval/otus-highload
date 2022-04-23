package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
	"strconv"
)

const (
	limitDefault  = 1000
	offsetDefault = 0

	offsetParameter = "offset"
	limitParameter  = "limit"
)

type Paging struct {
	Offset int
	Limit  int
}

func (p Paging) Next() Paging {
	return Paging{Offset: p.Offset + limitDefault, Limit: p.Limit + limitDefault}
}

func (p Paging) Previous() Paging {
	if p.Offset <= 0 {
		return Paging{Offset: 0, Limit: limitDefault}
	}
	return Paging{Offset: p.Offset - limitDefault, Limit: p.Limit - limitDefault}
}

func getPageParameters(controller *beego.Controller) (Paging, error) {
	paging := Paging{Offset: offsetDefault, Limit: limitDefault}
	var err error

	value := controller.GetString(offsetParameter)
	if value != "" {
		paging.Offset, err = strconv.Atoi(value)
		if err != nil {
			return Paging{}, err
		}
		if paging.Offset < 0 {
			paging.Offset = offsetDefault
		}
	}
	value = controller.GetString(limitParameter)
	if value != "" {
		paging.Limit, err = strconv.Atoi(value)
		if err != nil {
			return Paging{}, err
		}
		if paging.Limit < 0 {
			paging.Limit = limitDefault
		}
	}
	return paging, nil
}
