/**
 * @author      Liu Yongshuai
 * @date        2018-03-31 15:15
 */
package goUtils

import (
	"fmt"
	"github.com/kr/pretty"
	"os"
	"os/user"
	"path"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestToCBD(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	fmt.Printf("\n\n\n------------start %s------------\n", f.Name())
	str := "ａｂｃａ@￥@#%#ｓｄ🎈🎉ｆ我E２３４３４５んエォサ６３＃＄％＾＄＆％＾（＆我"
	fmt.Println(str)
	fmt.Println(ToCBD(str))
	fmt.Printf("------------end %s------------\n", f.Name())
}

func TestToDBC(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	fmt.Printf("\n\n\n------------start %s------------\n", f.Name())
	str := "んエォサ６３1234567sdgs sdfgsａｂ。......ｃａ@￥@#%#ｓｄ我"
	fmt.Println(str)
	fmt.Println(ToDBC(str))
	fmt.Printf("------------end %s------------\n", f.Name())
}

func TestLocalIP(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	fmt.Printf("\n\n\n------------start %s------------\n", f.Name())
	localIps := LocalIP()
	for _, ip := range localIps {
		fmt.Fprintf(os.Stdout, "localIP[%s] IsPrivate[%v]\n", ip, IsPrivateIP(ip))
	}
	user1, _ := user.Current()
	fmt.Println(user1.HomeDir)
	fmt.Printf("------------end %s------------\n", f.Name())
}

func TestPrintCallerName(t *testing.T) {
	PrintCallerName(0, "TestPrintCallerName")
}

// 获取调用者信息
func CallerName(skip int) (name, file string, line int, ok bool) {
	var (
		reInit    = regexp.MustCompile(`init·\d+$`) // main.init·1
		reClosure = regexp.MustCompile(`func·\d+$`) // main.func·001
	)
	for {
		var pc uintptr
		if pc, file, line, ok = runtime.Caller(skip + 1); !ok {
			return
		}
		name = runtime.FuncForPC(pc).Name()
		if reInit.MatchString(name) {
			name = reInit.ReplaceAllString(name, "init")
			return
		}
		if reClosure.MatchString(name) {
			skip++
			continue
		}
		return
	}
	return
}

// 输出调用者信息--调试使用
func PrintCallerName(skip int, comment string) (string, bool) {
	name, file, line, ok := CallerName(skip + 1)
	_, shortName := path.Split(name)
	if !ok {
		return shortName, false
	}
	fmt.Printf("\n===================================================\n")
	fmt.Printf("skip = %v, comment = %s\n", skip, comment)
	fmt.Printf("  file = %v, line = %d\n", file, line)
	fmt.Printf("  name = %v\n", name)
	return shortName, true
}

func TestIsNormalStr(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	fmt.Printf("\n\n\n------------start %s------------\n", f.Name())
	fmt.Println(IsNormalStr("ssss我&"))
	fmt.Println(IsNormalStr("馄饨面+wendao"))
	fmt.Println(IsNormalStr("面条1碗"))
	fmt.Printf("------------end %s------------\n", f.Name())
}

func TestRandFloat64(t *testing.T) {
	min, max := 39.67068, 41.060816
	for i := 0; i < 10000000; i++ {
		ret := RandFloat64InRange(min, max)
		if ret <= min || ret >= max {
			t.Errorf("Random number out of range : %f", ret)
		}
		//fmt.Printf("%f\n", ret)
	}
}

func TestStrHashSum64(t *testing.T) {
	s := "asdfasdfasdfasdf"
	fmt.Println(int64(StrHashSum64(s)))
}

func TestRandomStr(t *testing.T) {
	fmt.Println(RandomStr(15))
	fmt.Println(RandomStr(32))
}
func TestBase62(t *testing.T) {
	var i int64 = 349879
	b62 := Base62Encode(i)
	fmt.Println(b62)
	fmt.Println(Base62Decode(b62))
}

func TestPregReplaceCallback(t *testing.T) {
	originStr := `
<div class="dropdown global-dropdown">
	<button class="global-dropdown-toggle" data-toggle="dropdown" type="button">
		<span class="sr-only">Toggle navigation</span>
		<i aria-hidden='true' data-hidden = "true" class="fa fa-bars"></i>
	</button>
	<div class="dropdown-menu-nav global-dropdown-menu">
		<ul>
			<li class="home active">
				<a title="Projects" class="dashboard-shortcuts-projects" href="/dashboard/projects">
					<div class="shortcut-mappings">
						<div class="key">
							<i aria-label="hidden" class="fa fa-arrow-up"></i>
						</div>
					</div>
				</a>
			</li>
		</ul>
	</div>
</div>`
	//给所有的div标签加上一个属性
	regPattern1 := `<div(.*?)>`
	ss, _ := PregReplaceCallback(regPattern1, originStr, func(ms []string) string {
		//[]string{
		//   "<div class=\"dropdown global-dropdown\">",
		//   " class=\"dropdown global-dropdown\"",
		//}
		//[]string{
		// 	 "<div class=\"dropdown-menu-nav global-dropdown-menu\">",
		// 	 " class=\"dropdown-menu-nav global-dropdown-menu\"",
		//}
		//[]string{
		// 	 "<div class=\"shortcut-mappings\">",
		// 	 " class=\"shortcut-mappings\"",
		//}
		//[]string{
		// 	  "<div class=\"key\">",
		// 	  " class=\"key\"",
		//}
		//ms[0]是正则匹配的整个字符串，ms[1]表示正则中小括号捕获的子串
		fmt.Printf("%# v\n", pretty.Formatter(ms))
		return fmt.Sprintf("<div%s onclick=\"javascript:void(0);\">", ms[1])
	})
	fmt.Println("给所有的div标签加上一个属性", ss)
	//判断所有的i标签，如果包含data-hidden属性则添加另一个属性
	regPattern2 := `<i(.*?)>`
	ss, _ = PregReplaceCallback(regPattern2, originStr, func(ms []string) string {
		if len(ms) < 1 {
			return ms[0]
		}
		iattr := strings.TrimSpace(ms[1])
		//将各个属性切开，分隔符取引号后跟空格
		reg1, _ := regexp.Compile(`["|']\s+`)
		attrArr := reg1.Split(iattr, -1)
		for _, attr := range attrArr {
			//将各个属性切开，注意"="等号左右有可能有空格
			reg2, _ := regexp.Compile(`\s*=\s*`)
			tmpArr := reg2.Split(attr, -1)
			if len(tmpArr) != 2 {
				continue
			}
			if tmpArr[0] == "data-hidden" {
				return fmt.Sprintf("<i%s selfAttr=\"1\">", ms[1])
			}
		}
		return ms[0]
	})
	fmt.Println("修改了i标签的属性", ss)
}

func TestOpenNewFile(t *testing.T) {
	f := "/Users/liuyongshuai/Documents/wendao/liu/sss/asdfasdfdsaf/abc.txt"
	fp, err := OpenNewFile(f, "", true)
	fmt.Println(err)
	fp.Close()
}

func TestTryBestConvert(t *testing.T) {
	p1 := 45649065094658405684504232323223322334.555
	p2 := "45s89s"
	p3 := "wendao"
	p4 := &p2
	vals := []interface{}{
		"34343434",
		44.3222,
		989889,
		0.222,
		&p1,
		&p2,
		&p3,
		&p4,
		"",
		true,
		-22222,
	}
	for _, val := range vals {
		int64Val, int64Err := TryBestToInt64(val)
		uint64Val, uint64Err := TryBestToUint64(val)
		floatVal, floatErr := TryBestToFloat(val)
		strVal, strErr := TryBestToString(val)
		boolVal, boolErr := TryBestToBool(val)
		fmt.Printf("rawVal %# v \tint64[%v %v] uint64[%v %v] float[%v %v] str[%v %v] bool[%v %v]\n",
			pretty.Formatter(val),
			pretty.Formatter(int64Val), int64Err,
			pretty.Formatter(uint64Val), uint64Err,
			pretty.Formatter(floatVal), floatErr,
			pretty.Formatter(strVal), strErr,
			pretty.Formatter(boolVal), boolErr,
		)
	}
}

func TestFilterIds(t *testing.T) {
	ids := []interface{}{
		3434,
		-9999,
		"34343443",
	}
	ret := FilterIds(ids)
	fmt.Println(ret)
}

func TestPrintTextDiff(t *testing.T) {
	text1 := `
45454545454545454  特朗普谈美国向移民发射催泪弹：移民很粗暴
sadfadsad 安徽：助力民企发展壮大 支持民营企业在行动
xcvxcvxc 特朗普喊话中美洲移民:如有必要 将永久关闭边境
sss 湖南隆回暂缓"百元车位" 中标单位曾被指可获暴利
】=-、意媒谈D&G风波：中国人记性差 抵制不了多久
*&）……（&暖新闻 带脑瘫儿子跑马拉松 父亲:让儿子少留遗憾
`
	text2 := `
、1222！@￥洞察"号登陆火星传首张照片:可见火星地平线
）（&**&……&……￥%￥##！日本茨城县发生5级地震多县有震感 尚未引发海啸
女子被顶替上学?堂姐夫:她考前已去卖猪肉 没考试
礌lklasdjgfakldgja5岁儿童简历长15页 人民日报:拔苗种不出好"庄稼"
2135457950875607网红自称回深山卖土蜂蜜 所留地址村委会:无此人
||||||||||\、、、、、、暴风雪袭击美国芝加哥地区 近900个航班被取消
`
	PrintTextDiff(text1, text2)
}

func TestPrintDiffTextByGroup(t *testing.T) {
	leftText := [][]string{
		{
			"45454545454545454  特朗普谈美国向移民发射催泪弹：移民很粗暴",
			"sadfadsad 安徽：助力民企发展壮大 支持民营企业在行动",
		},
		{
			"xcvxcvxc 特朗普喊话中美洲移民:如有必要 将永久关闭边境",
		},
		{
			"45454545454545454  特朗普谈美国向移民发射催泪弹：移民很粗暴",
			"sadfadsad 安徽：助力民企发展壮大 支持民营企业在行动",
		},
		{
			"xcvxcvxc 特朗普喊话中美洲移民:如有必要 将永久关闭边境",
		},
	}
	rightText := [][]string{
		{
			"、1222！@￥洞察号登陆火星传首张照片:可见火星地平线",
		},
		{
			"xcvxcvxc 特朗普喊话中美洲移民:如有必要 将永久关闭边境",
		},
		{
			"）（&**&……&……￥%￥##！日本茨城县发生5级地震多县有震感 尚未引发海啸",
			"女子被顶替上学?堂姐夫:她考前已去卖猪肉 没考试",
		},
		{
			"、1222！@￥洞察号登陆火星传首张照片:可见火星地平线",
			"）（&**&……&……￥%￥##！日本茨城县发生5级地震多县有震感 尚未引发海啸",
			"女子被顶替上学?堂姐夫:她考前已去卖猪肉 没考试",
			"礌lklasdjgfakldgja5岁儿童简历长15页 人民日报:拔苗种不出好庄稼",
			"2135457950875607网红自称回深山卖土蜂蜜 所留地址村委会:无此人",
			"||||||||||、、、、、、暴风雪袭击美国芝加哥地区 近900个航班被取消",
		},
	}
	PrintTextDiffByGroup(leftText, rightText)
	PrintTextDiffByGroup(leftText, [][]string{})
	PrintTextDiffByGroup([][]string{}, rightText)
}

func TestWrap(t *testing.T) {
	str := "暴风雪袭击美国芝加哥地区"
	fmt.Println(RuneWrap(str, 7))
	/**
暴风雪
袭击美
国芝加
哥地区
	 */
}
