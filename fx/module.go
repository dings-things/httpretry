package fx

import (
	"github.com/dings-things/httpretry"
	"go.uber.org/fx"
)

// Module httpretry 클라이언트의 고정 종속성을 제공합니다
var Module = fx.Module(
	"dings-things/httpretry",
	fx.Provide(
		// retry client
		fx.Annotate(httpretry.NewClient, fx.ResultTags(`name:"httpclient"`)),

		httpretry.NewSettings,
	),
)
