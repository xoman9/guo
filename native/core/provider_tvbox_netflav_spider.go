package core

import (
    "context"
    "errors"
    "regexp"
)

// Netflav is rendered by the Android TVBox GMSpider JavaScript bridge.
// Register in Android's Go core so the Flutter source picker can find it,
// while making the unsupported native-only path explicit.
const tvboxNetflavSource = "tvbox-netflav"
var tvboxNetflavID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,160}$`)

type tvboxNetflavJSProvider struct{}
func (*tvboxNetflavJSProvider) Descriptor() SourceDescriptor {
    return SourceDescriptor{
        ID:tvboxNetflavSource,Name:"Netflav",DefaultOff:true,
        BaseURL:"https://netflav.com",NoHostRotation:true,
        Description:"使用內建 TVBox GMSpider JavaScript 與 Android WebView（Android 專用）",
        ValidCategory:func(string)bool{return true},
    }
}
var errNetflavJSRequired = errors.New("Netflav 需要 Android TVBox JavaScript 執行器")
func (*tvboxNetflavJSProvider) Categories(context.Context,bool)([]Category,error) {
    return nil,errNetflavJSRequired
}
func (*tvboxNetflavJSProvider) CatalogPage(context.Context,int,string)([]Drama,bool,error) {
    return nil,false,errNetflavJSRequired
}
func (*tvboxNetflavJSProvider) Search(context.Context,int,string)([]Drama,bool,error) {
    return nil,false,errNetflavJSRequired
}
func (*tvboxNetflavJSProvider) Detail(context.Context,string)(Drama,[]Chapter,error) {
    return Drama{},nil,errNetflavJSRequired
}
func (*tvboxNetflavJSProvider) ResolveMedia(context.Context,Task)(ProviderMedia,error) {
    return ProviderMedia{},errNetflavJSRequired
}
func init(){
    descriptor:=(&tvboxNetflavJSProvider{}).Descriptor()
    Register(descriptor,func(Host)Provider{return &tvboxNetflavJSProvider{}})
}
