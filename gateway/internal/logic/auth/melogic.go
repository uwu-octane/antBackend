// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

<<<<<<<< HEAD:gateway/internal/logic/api/loginlogic.go
package api
========
package auth
>>>>>>>> integrationAuth:gateway/internal/logic/auth/melogic.go

import (
	"context"

	"github.com/uwu-octane/antBackend/gateway/internal/svc"
	"github.com/uwu-octane/antBackend/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

<<<<<<<< HEAD:gateway/internal/logic/api/loginlogic.go
type LoginLogic struct {
========
type MeLogic struct {
>>>>>>>> integrationAuth:gateway/internal/logic/auth/melogic.go
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

<<<<<<<< HEAD:gateway/internal/logic/api/loginlogic.go
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
========
func NewMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MeLogic {
	return &MeLogic{
>>>>>>>> integrationAuth:gateway/internal/logic/auth/melogic.go
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

<<<<<<<< HEAD:gateway/internal/logic/api/loginlogic.go
func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
========
func (l *MeLogic) Me() (resp *types.MeResp, err error) {
>>>>>>>> integrationAuth:gateway/internal/logic/auth/melogic.go
	// todo: add your logic here and delete this line

	return
}
