package config

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gview"
	"github.com/gogf/gf/v2/text/gstr"
)

var (
	CHATPROXY    = "http://demo.xyhelper.cn"
	AUTHKEY      = "xyhelper"
	ArkoseUrl    = "https://tcr9i.closeai.biz/v2/"
	BuildId      = "q1uPMR9VmRS6HPYxPKMbH"
	CacheBuildId = "q1uPMR9VmRS6HPYxPKMbH"
	AssetPrefix  = "https://oaistatic-cdn.closeai.biz"
	PK40         = "35536E1E-65B4-4D96-9D97-6ADB7EFF8147"
	PK35         = "3D86FBBA-9D22-402A-B512-3420086BA6CC"
	envScriptTpl = `
	<script>
	window.__arkoseUrl="{{.ArkoseUrl}}";
	window.__assetPrefix="{{.AssetPrefix}}";
	window.__PK40="{{.PK40}}";
	window.__PK35="{{.PK35}}";
	</script>
	`
)

func init() {
	ctx := gctx.GetInitCtx()
	chatproxy := g.Cfg().MustGetWithEnv(ctx, "CHATPROXY").String()
	if chatproxy != "" {
		CHATPROXY = chatproxy
	}
	g.Log().Info(ctx, "CHATPROXY:", CHATPROXY)
	authkey := g.Cfg().MustGetWithEnv(ctx, "AUTHKEY").String()
	if authkey != "" {
		AUTHKEY = authkey
	}
	g.Log().Info(ctx, "AUTHKEY:", AUTHKEY)
	arkoseUrl := g.Cfg().MustGetWithEnv(ctx, "ARKOSE_URL")
	if !arkoseUrl.IsEmpty() {
		ArkoseUrl = arkoseUrl.String()
	}
	assetPrefix := g.Cfg().MustGetWithEnv(ctx, "ASSET_PREFIX").String()
	if assetPrefix != "" {
		AssetPrefix = assetPrefix
	}
	g.Log().Info(ctx, "ASSET_PREFIX:", AssetPrefix)
	cacheBuildId := CheckVersion(ctx, AssetPrefix)
	if cacheBuildId != "" {
		CacheBuildId = cacheBuildId
	}
	g.Log().Info(ctx, "CacheBuildId:", CacheBuildId)
	build := CheckNewVersion(ctx)
	if build != "" {
		BuildId = build
	}
	g.Log().Info(ctx, "BuildId:", BuildId)
}

func GetEnvScript(ctx g.Ctx) string {
	script, err := gview.ParseContent(ctx, envScriptTpl, g.Map{
		"ArkoseUrl":   ArkoseUrl,
		"AssetPrefix": AssetPrefix,
		"PK40":        PK40,
		"PK35":        PK35,
	})
	if err != nil {
		g.Log().Error(ctx, "GetEnvScript Error: ", err)
		return ""
	}
	return script
}

// 检查是否有新版本
func CheckNewVersion(ctx g.Ctx) (buildId string) {
	resVar := g.Client().GetVar(ctx, CHATPROXY+"/ping")
	resJson := gjson.New(resVar)

	buildId = resJson.Get("buildId").String()
	return
}

// 检查版本号并同步资源
func CheckVersion(ctx g.Ctx, assetPrefix string) (CacheBuildId string) {
	gclient := g.Client()
	// 读取 assetPrefix/version
	versionVar := gclient.GetVar(ctx, assetPrefix+"/version.json")
	CacheBuildId = gjson.New(versionVar).Get("cacheBuildId").String()
	g.Log().Infof(ctx, "Get config From %s ,CacheBuildId: %s", AssetPrefix, CacheBuildId)
	if CacheBuildId == "" {
		return ""
	}
	// 读取buildDate目录索引
	indexUrl := assetPrefix + "/template/" + CacheBuildId + "/index.txt"
	g.Log().Info(ctx, "Get config From ", indexUrl)
	buildDateVar := gclient.GetVar(ctx, indexUrl).String()
	if buildDateVar == "" {
		return ""
	}
	// 按回车分割
	buildDateList := gstr.Split(buildDateVar, "\n")
	// g.Dump(buildDateList)
	// 遍历目录索引 如果没有就下载
	for _, v := range buildDateList {
		if v == "" {
			continue
		}
		// 检查文件是否存在
		if !gfile.Exists("./resource/template/" + CacheBuildId + "/" + v) {
			g.Log().Infof(ctx, "Download %s", v)
			// 下载文件
			res, err := gclient.Get(ctx, assetPrefix+"/template/"+CacheBuildId+"/"+v)
			if err != nil {
				g.Log().Error(ctx, "Download  Error: ", v, err)
				return ""
			}
			defer res.Close()
			if res.StatusCode != 200 {
				g.Log().Error(ctx, "Download  Error: ", v, res.StatusCode)
				return ""
			}
			// 写入文件
			err = gfile.PutBytes("./resource/template/"+CacheBuildId+"/"+v, res.ReadAll())
			if err != nil {
				g.Log().Error(ctx, "Download  Error: ", v, err)
				return ""
			}

		}
	}

	return
}