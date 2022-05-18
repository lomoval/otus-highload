package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

	beego.GlobalControllerRouter["app/controllers:DialogController"] = append(beego.GlobalControllerRouter["app/controllers:DialogController"],
		beego.ControllerComments{
			Method:           "Dialogs",
			Router:           "/dialogs",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["app/controllers:DialogController"] = append(beego.GlobalControllerRouter["app/controllers:DialogController"],
		beego.ControllerComments{
			Method:           "DialogAnswers",
			Router:           "/dialogs/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams: param.Make(
				param.New("id", param.InPath),
			),
			Filters: nil,
			Params:  nil})

	beego.GlobalControllerRouter["app/controllers:NewsController"] = append(beego.GlobalControllerRouter["app/controllers:NewsController"],
		beego.ControllerComments{
			Method:           "GetFriendsNews",
			Router:           "/friends/news",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["app/controllers:NewsController"] = append(beego.GlobalControllerRouter["app/controllers:NewsController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           "/news",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["app/controllers:NewsController"] = append(beego.GlobalControllerRouter["app/controllers:NewsController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/news/",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["app/controllers:ProfileController"] = append(beego.GlobalControllerRouter["app/controllers:ProfileController"],
		beego.ControllerComments{
			Method:           "AddDialog",
			Router:           "/dialogs",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["app/controllers:ProfileController"] = append(beego.GlobalControllerRouter["app/controllers:ProfileController"],
		beego.ControllerComments{
			Method:           "AddDialogAnswer",
			Router:           "/dialogs/:id",
			AllowHTTPMethods: []string{"post"},
			MethodParams: param.Make(
				param.New("id", param.InPath),
			),
			Filters: nil,
			Params:  nil})

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
