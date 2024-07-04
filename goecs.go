package main

import (
	"flag"
	"fmt"
	"github.com/oneclickvirt/CommonMediaTests/commediatests"
	unlocktestmodel "github.com/oneclickvirt/UnlockTests/model"
	backtraceori "github.com/oneclickvirt/backtrace/bk"
	basicmodel "github.com/oneclickvirt/basics/model"
	cputestmodel "github.com/oneclickvirt/cputest/model"
	disktestmodel "github.com/oneclickvirt/disktest/disk"
	"github.com/oneclickvirt/ecs/backtrace"
	"github.com/oneclickvirt/ecs/commediatest"
	"github.com/oneclickvirt/ecs/cputest"
	"github.com/oneclickvirt/ecs/disktest"
	"github.com/oneclickvirt/ecs/memorytest"
	"github.com/oneclickvirt/ecs/ntrace"
	"github.com/oneclickvirt/ecs/speedtest"
	"github.com/oneclickvirt/ecs/unlocktest"
	"github.com/oneclickvirt/ecs/utils"
	gostunmodel "github.com/oneclickvirt/gostun/model"
	memorytestmodel "github.com/oneclickvirt/memorytest/memory"
	nt3model "github.com/oneclickvirt/nt3/model"
	ptmodel "github.com/oneclickvirt/pingtest/model"
	"github.com/oneclickvirt/pingtest/pt"
	"github.com/oneclickvirt/portchecker/email"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	ecsVersion                                                        = "v0.0.48"
	menuMode                                                          bool
	onlyChinaTest                                                     bool
	input, choice                                                     string
	showVersion                                                       bool
	enableLogger                                                      bool
	language                                                          string
	cpuTestMethod, cpuTestThreadMode                                  string
	memoryTestMethod                                                  string
	diskTestMethod, diskTestPath                                      string
	diskMultiCheck                                                    bool
	nt3CheckType, nt3Location                                         string
	spNum                                                             int
	width                                                             = 82
	basicStatus, cpuTestStatus, memoryTestStatus, diskTestStatus      bool
	commTestStatus, utTestStatus, securityTestStatus, emailTestStatus bool
	backtraceStatus, nt3Status, speedTestStatus, pingTestStatus       bool
	filePath                                                          = "goecs.txt"
	enabelUpload                                                      = true
	help                                                              bool
	goecsFlag                                                         = flag.NewFlagSet("goecs", flag.ContinueOnError)
)

func main() {
	goecsFlag.BoolVar(&help, "h", false, "Show help information")
	goecsFlag.BoolVar(&showVersion, "v", false, "Display version information")
	goecsFlag.BoolVar(&menuMode, "menu", true, "Enable/Disable menu mode, disable example: -menu=false") // true 默认启用菜单栏模式
	goecsFlag.StringVar(&language, "l", "zh", "Set language (supported: en, zh)")
	goecsFlag.BoolVar(&basicStatus, "basic", true, "Enable/Disable basic test")
	goecsFlag.BoolVar(&cpuTestStatus, "cpu", true, "Enable/Disable CPU test")
	goecsFlag.BoolVar(&memoryTestStatus, "memory", true, "Enable/Disable memory test")
	goecsFlag.BoolVar(&diskTestStatus, "disk", true, "Enable/Disable disk test")
	goecsFlag.BoolVar(&commTestStatus, "comm", true, "Enable/Disable common media test")
	goecsFlag.BoolVar(&utTestStatus, "ut", true, "Enable/Disable unlock media test")
	goecsFlag.BoolVar(&securityTestStatus, "security", true, "Enable/Disable security test")
	goecsFlag.BoolVar(&emailTestStatus, "email", true, "Enable/Disable email port test")
	goecsFlag.BoolVar(&backtraceStatus, "backtrace", true, "Enable/Disable backtrace test (in 'en' language or on windows it always false)")
	goecsFlag.BoolVar(&nt3Status, "nt3", true, "Enable/Disable NT3 test (in 'en' language or on windows it always false)")
	goecsFlag.BoolVar(&speedTestStatus, "speed", true, "Enable/Disable speed test")
	goecsFlag.StringVar(&cpuTestMethod, "cpum", "sysbench", "Set CPU test method (supported: sysbench, geekbench, winsat)")
	goecsFlag.StringVar(&cpuTestThreadMode, "cput", "multi", "Set CPU test thread mode (supported: single, multi)")
	goecsFlag.StringVar(&memoryTestMethod, "memorym", "dd", "Set memory test method (supported: sysbench, dd, winsat)")
	goecsFlag.StringVar(&diskTestMethod, "diskm", "fio", "Set disk test method (supported: fio, dd, winsat)")
	goecsFlag.StringVar(&diskTestPath, "diskp", "", "Set disk test path, e.g., -diskp /root")
	goecsFlag.BoolVar(&diskMultiCheck, "diskmc", false, "Enable/Disable multiple disk checks, e.g., -diskmc=false")
	goecsFlag.StringVar(&nt3Location, "nt3loc", "GZ", "Specify NT3 test location (supported: GZ, SH, BJ, CD for Guangzhou, Shanghai, Beijing, Chengdu)")
	goecsFlag.StringVar(&nt3CheckType, "nt3t", "ipv4", "Set NT3 test type (supported: both, ipv4, ipv6)")
	goecsFlag.IntVar(&spNum, "spnum", 2, "Set the number of servers per operator for speed test")
	goecsFlag.BoolVar(&enableLogger, "log", false, "Enable/Disable logging in the current path")
	goecsFlag.BoolVar(&enabelUpload, "upload", true, "Enable/Disable upload the result")
	goecsFlag.Parse(os.Args[1:])
	if help {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		goecsFlag.PrintDefaults()
		return
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	if showVersion {
		fmt.Println(ecsVersion)
		return
	}
	if enableLogger {
		gostunmodel.EnableLoger = true
		basicmodel.EnableLoger = true
		cputestmodel.EnableLoger = true
		memorytestmodel.EnableLoger = true
		disktestmodel.EnableLoger = true
		commediatests.EnableLoger = true
		unlocktestmodel.EnableLoger = true
		ptmodel.EnableLoger = true
		backtraceori.EnableLoger = true
		nt3model.EnableLoger = true
		speedtestmodel.EnableLoger = true
	}
	go func() {
		http.Get("https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Foneclickvirt%2Fecs&count_bg=%2357DEFF&title_bg=%23000000&icon=cliqz.svg&icon_color=%23E7E7E7&title=hits&edge_flat=false")
	}()
	if menuMode {
		basicStatus, cpuTestStatus, memoryTestStatus, diskTestStatus = false, false, false, false
		commTestStatus, utTestStatus, securityTestStatus, emailTestStatus = false, false, false, false
		backtraceStatus, nt3Status, speedTestStatus = false, false, false
		switch language {
		case "zh":
			fmt.Println("VPS融合怪版本: ", ecsVersion)
			fmt.Println("1. 融合怪完全体")
			fmt.Println("2. 极简版(系统信息+CPU+内存+磁盘+测速节点5个)")
			fmt.Println("3. 精简版(系统信息+CPU+内存+磁盘+御三家+常用流媒体+回程+路由+测速节点5个)")
			fmt.Println("4. 精简网络版(系统信息+CPU+内存+磁盘+回程+路由+测速节点5个)")
			fmt.Println("5. 精简解锁版(系统信息+CPU+内存+磁盘IO+御三家+常用流媒体+测速节点5个)")
			fmt.Println("6. 网络单项(IP质量检测+三网回程+三网路由与延迟+测速节点11个)")
			fmt.Println("7. 解锁单项(御三家解锁+常用流媒体解锁)")
			fmt.Println("8. 硬件单项(基础系统信息+CPU+内存+dd磁盘测试+fio磁盘测试)")
			fmt.Println("9. IP质量检测(15个数据库的IP检测+邮件端口检测)")
			fmt.Println("10. 三网回程线路+广州三网路由+全国三网延迟")
		case "en":
			fmt.Println("VPS Fusion Monster Test Version: ", ecsVersion)
			fmt.Println("1. VPS Fusion Monster Test Comprehensive Test Suite")
			fmt.Println("2. Minimal Test Suite (System Info + CPU + Memory + Disk + 5 Speed Test Nodes)")
			fmt.Println("3. Standard Test Suite (System Info + CPU + Memory + Disk + Basic Unlock Tests + Common Streaming Services + 5 Speed Test Nodes)")
			fmt.Println("4. Network-Focused Test Suite (System Info + CPU + Memory + Disk + 5 Speed Test Nodes)")
			fmt.Println("5. Unlock-Focused Test Suite (System Info + CPU + Memory + Disk IO + Basic Unlock Tests + Common Streaming Services + 5 Speed Test Nodes)")
			fmt.Println("6. Network-Only Test (IP Quality Test + 5 Speed Test Nodes)")
			fmt.Println("7. Unlock-Only Test (Basic Unlock Tests + Common Streaming Services Unlock)")
			fmt.Println("8. Hardware-Only Test (Basic System Info + CPU + Memory + dd Disk Test + fio Disk Test)")
			fmt.Println("9. IP Quality Test (IP Test with 15 Databases + Email Port Test)")
		}
	Loop:
		for {
			fmt.Print("请输入选项 / Please enter your choice: ")
			fmt.Scanln(&input)
			input = strings.TrimSpace(input)
			input = strings.TrimRight(input, "\n")
			re := regexp.MustCompile(`^\d+$`) // 正则表达式匹配纯数字
			if re.MatchString(input) {
				choice = input
				switch choice {
				case "1":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					commTestStatus = true
					utTestStatus = true
					securityTestStatus = true
					emailTestStatus = true
					backtraceStatus = true
					nt3Status = true
					speedTestStatus = true
					onlyChinaTest = utils.CheckChina(enableLogger)
					break Loop
				case "2":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					speedTestStatus = true
					break Loop
				case "3":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					commTestStatus = true
					utTestStatus = true
					securityTestStatus = true
					backtraceStatus = true
					nt3Status = true
					speedTestStatus = true
					break Loop
				case "4":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					backtraceStatus = true
					nt3Status = true
					speedTestStatus = true
					break Loop
				case "5":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					securityTestStatus = true
					speedTestStatus = true
					break Loop
				case "6":
					speedTestStatus = true
					backtraceStatus = true
					nt3Status = true
					break Loop
				case "7":
					securityTestStatus = true
					commTestStatus = true
					break Loop
				case "8":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					break Loop
				case "9":
					emailTestStatus = true
					break Loop
				case "10":
					backtraceStatus = true
					nt3Status = true
					speedTestStatus = true
					pingTestStatus = true
					break Loop
				default:
					if language == "zh" {
						fmt.Println("无效的选项")
					} else {
						fmt.Println("Invalid choice")
					}
				}
			} else {
				if language == "zh" {
					fmt.Println("输入错误，请输入一个纯数字")
				} else {
					fmt.Println("Invalid input, please enter a number")
				}
			}
		}
	}
	if language == "en" {
		backtraceStatus = false
		nt3Status = false
	}
	if !enabelUpload {
		securityTestStatus = false
	}
	var (
		startTime                                     time.Time
		wg1, wg2                                      sync.WaitGroup
		basicInfo, securityInfo, emailInfo, mediaInfo string
		output, tempOutput                            string
	)
	// 启动一个goroutine来等待信号，内置计时器
	go func() {
		startTime = time.Now()
		// 等待信号
		<-sig
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
		output = utils.PrintAndCapture(func() {
			utils.PrintCenteredTitle("", width)
			fmt.Printf("Cost    Time          : %d min %d sec\n", minutes, seconds)
			fmt.Printf("Current Time          : %s\n", currentTime)
			utils.PrintCenteredTitle("", width)
		}, tempOutput, output)
		utils.ProcessAndUpload(output, filePath, enabelUpload)
		os.Exit(1) // 使用非零状态码退出，表示意外退出
	}()
	switch language {
	case "zh":
		output = utils.PrintAndCapture(func() {
			utils.PrintHead(language, width, ecsVersion)
			if basicStatus || securityTestStatus {
				if basicStatus {
					utils.PrintCenteredTitle("系统基础信息", width)
				}
				basicInfo, securityInfo, nt3CheckType = utils.SecurityCheck(language, nt3CheckType, securityTestStatus)
				if basicStatus {
					fmt.Printf(basicInfo)
				}
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if cpuTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", cpuTestMethod), width)
				cputest.CpuTest(language, cpuTestMethod, cpuTestThreadMode)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if memoryTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", cpuTestMethod), width)
				memorytest.MemoryTest(language, memoryTestMethod)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if diskTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", diskTestMethod), width)
				disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
			}
		}, tempOutput, output)
		if emailTestStatus {
			wg2.Add(1)
			go func() {
				defer wg2.Done()
				emailInfo = email.EmailCheck()
			}()
		}
		if utTestStatus && !onlyChinaTest {
			wg1.Add(1)
			go func() {
				defer wg1.Done()
				mediaInfo = unlocktest.MediaTest(language)
			}()
		}
		output = utils.PrintAndCapture(func() {
			if commTestStatus && !onlyChinaTest {
				utils.PrintCenteredTitle("御三家流媒体解锁", width)
				commediatest.ComMediaTest(language)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if utTestStatus && !onlyChinaTest {
				utils.PrintCenteredTitle("跨国流媒体解锁", width)
				wg1.Wait()
				fmt.Printf(mediaInfo)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if securityTestStatus {
				utils.PrintCenteredTitle("IP质量检测", width)
				fmt.Printf(securityInfo)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if emailTestStatus {
				utils.PrintCenteredTitle("邮件端口检测", width)
				wg2.Wait()
				fmt.Println(emailInfo)
			}
		}, tempOutput, output)
		if runtime.GOOS != "windows" {
			output = utils.PrintAndCapture(func() {
				if backtraceStatus && !onlyChinaTest {
					utils.PrintCenteredTitle("三网回程线路检测", width)
					backtrace.BackTrace()
				}
			}, tempOutput, output)
			// nexttrace 在win上不支持检测，报错 bind: An invalid argument was supplied.
			output = utils.PrintAndCapture(func() {
				if nt3Status && !onlyChinaTest {
					utils.PrintCenteredTitle("三网回程路由检测", width)
					ntrace.TraceRoute3(language, nt3Location, nt3CheckType)
				}
			}, tempOutput, output)
			output = utils.PrintAndCapture(func() {
				if onlyChinaTest || pingTestStatus {
					utils.PrintCenteredTitle("三网ICMP的PING值检测", width)
					pt.PingTest()
				}
			}, tempOutput, output)
		}
		output = utils.PrintAndCapture(func() {
			if speedTestStatus {
				utils.PrintCenteredTitle("就近节点测速", width)
				speedtest.ShowHead(language)
				if (menuMode && choice == "1") || !menuMode {
					speedtest.NearbySP()
					speedtest.CustomSP("net", "global", 2, language)
					speedtest.CustomSP("net", "cu", spNum, language)
					speedtest.CustomSP("net", "ct", spNum, language)
					speedtest.CustomSP("net", "cmcc", spNum, language)
				} else if menuMode && choice == "2" || choice == "3" || choice == "4" || choice == "5" {
					speedtest.CustomSP("net", "global", 4, language)
				}
			}
		}, tempOutput, output)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
		output = utils.PrintAndCapture(func() {
			utils.PrintCenteredTitle("", width)
			fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
			fmt.Printf("时间          : %s\n", currentTime)
			utils.PrintCenteredTitle("", width)
		}, tempOutput, output)
	case "en":
		output = utils.PrintAndCapture(func() {
			utils.PrintHead(language, width, ecsVersion)
			if basicStatus || securityTestStatus {
				if basicStatus {
					utils.PrintCenteredTitle("System-Basic-Information", width)
				}
				basicInfo, securityInfo, nt3CheckType = utils.SecurityCheck(language, nt3CheckType, securityTestStatus)
				if basicStatus {
					fmt.Printf(basicInfo)
				}
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if cpuTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU-Test--%s-Method", cpuTestMethod), width)
				cputest.CpuTest(language, cpuTestMethod, cpuTestThreadMode)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if memoryTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("Memory-Test--%s-Method", memoryTestMethod), width)
				memorytest.MemoryTest(language, memoryTestMethod)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if diskTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("Disk-Test--%s-Method", diskTestMethod), width)
				disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
			}
		}, tempOutput, output)
		if utTestStatus {
			wg1.Add(1)
			go func() {
				defer wg1.Done()
				mediaInfo = unlocktest.MediaTest(language)
			}()
		}
		if emailTestStatus {
			wg2.Add(1)
			go func() {
				defer wg2.Done()
				emailInfo = email.EmailCheck()
			}()
		}
		output = utils.PrintAndCapture(func() {
			if utTestStatus {
				utils.PrintCenteredTitle("Cross-Border-Streaming-Media-Unlock", width)
				wg1.Wait()
				fmt.Printf(mediaInfo)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if securityTestStatus {
				utils.PrintCenteredTitle("IP-Quality-Check", width)
				fmt.Printf(securityInfo)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if emailTestStatus {
				utils.PrintCenteredTitle("Email-Port-Check", width)
				wg2.Wait()
				fmt.Println(emailInfo)
			}
		}, tempOutput, output)
		output = utils.PrintAndCapture(func() {
			if speedTestStatus {
				utils.PrintCenteredTitle("Speed-Test", width)
				speedtest.ShowHead(language)
				speedtest.NearbySP()
				speedtest.CustomSP("net", "global", -1, language)
			}
		}, tempOutput, output)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
		output = utils.PrintAndCapture(func() {
			utils.PrintCenteredTitle("", width)
			fmt.Printf("Cost    Time          : %d min %d sec\n", minutes, seconds)
			fmt.Printf("Current Time          : %s\n", currentTime)
			utils.PrintCenteredTitle("", width)
		}, tempOutput, output)
	default:
		fmt.Println("Unsupported language")
	}
	utils.ProcessAndUpload(output, filePath, enabelUpload)
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
