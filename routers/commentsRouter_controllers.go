package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

	beego.GlobalControllerRouter["app/controllers:ProfileController"] = append(beego.GlobalControllerRouter["app/controllers:ProfileController"],
		beego.ControllerComments{
			Method:           "Profile",
			Router:           "/profile",
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(
				param.New("id"),
			),
			Filters: nil,
			Params:  nil})

	beego.GlobalControllerRouter["app/controllers:ProfileController"] = append(beego.GlobalControllerRouter["app/controllers:ProfileController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/profile",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
