package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	all := flag.Bool("all", false, "列出所有支持的语言")
	java := flag.Bool("java", false, "Java 版本管理入口")
	javalist := flag.Bool("javalist", false, "兼容参数：列出支持的 Java 版本")
	javainstall := flag.String("javainstall", "", "兼容参数：安装指定 Java 主版本（例如 17）")
	javauninstall := flag.String("javauninstall", "", "兼容参数：卸载指定 Java 主版本（例如 17）")
	javause := flag.String("javause", "", "兼容参数：切换到指定 Java 主版本（例如 17）")
	javaList := flag.Bool("javaList", false, "兼容旧参数：列出支持的 Java 版本")
	javaInstall := flag.String("javaInstall", "", "兼容旧参数：安装指定 Java 主版本")
	javaUninstall := flag.String("javaUninstall", "", "兼容旧参数：卸载指定 Java 主版本")
	javaUse := flag.String("javaUse", "", "兼容旧参数：切换到指定 Java 主版本")

	python := flag.Bool("python", false, "Python 版本管理入口")
	pythonlist := flag.Bool("pythonlist", false, "兼容参数：列出支持的 Python 版本")
	pythoninstall := flag.String("pythoninstall", "", "兼容参数：安装指定 Python 主版本（例如 3.12）")
	pythonuninstall := flag.String("pythonuninstall", "", "兼容参数：卸载指定 Python 主版本（例如 3.12）")
	pythonuse := flag.String("pythonuse", "", "兼容参数：切换到指定 Python 主版本（例如 3.12）")

	// 通用子命令（与语言入口组合使用）：-list/-install/-uninstall/-use
	list := flag.Bool("list", false, "列出支持的版本")
	install := flag.String("install", "", "安装指定版本")
	uninstall := flag.String("uninstall", "", "卸载指定版本")
	use := flag.String("use", "", "切换到指定版本（生成 shim）")

	node := flag.Bool("node", false, "Node.js 版本管理入口")
	golang := flag.Bool("go", false, "Go 版本管理入口")
	rust := flag.Bool("rust", false, "Rust 版本管理入口")
	doctor := flag.Bool("doctor", false, "环境自检")

	flag.Parse()

	if *all {
		listSupportedLangs()
		return
	}

	// 不允许同时选择多个语言入口
	selected := 0
	for _, v := range []bool{*java, *python, *node, *golang, *rust} {
		if v {
			selected++
		}
	}
	if selected > 1 {
		fmt.Println("请选择一个语言入口：-java/-python/-node/-go/-rust；或使用 -all。")
		return
	}
	if selected == 0 && !*all && !*doctor {
		fmt.Println("请指定语言入口：-java/-python/-node/-go/-rust；或使用 -all。")
		return
	}

	if *doctor {
		runDoctor()
		return
	}

	if *java {
		// 若仅指定 -java，不带其它子参数，则输出当前 Java 版本
		if !(*list || *javalist || *javaList) && (chooseStr(*install, chooseStr(*javainstall, *javaInstall)) == "") && (chooseStr(*uninstall, chooseStr(*javauninstall, *javaUninstall)) == "") && (chooseStr(*use, chooseStr(*javause, *javaUse)) == "") {
			printCurrentJavaVersion()
			return
		}
		handleJava((*list || *javalist || *javaList), chooseStr(*install, chooseStr(*javainstall, *javaInstall)), chooseStr(*uninstall, chooseStr(*javauninstall, *javaUninstall)), chooseStr(*use, chooseStr(*javause, *javaUse)))
		return
	}

	if *python {
		if !(*list || *pythonlist) && (*install == "" && *pythoninstall == "") && (*uninstall == "" && *pythonuninstall == "") && (*use == "" && *pythonuse == "") {
			printCurrentPythonVersion()
			return
		}
		handlePython((*list || *pythonlist), chooseStr(*install, *pythoninstall), chooseStr(*uninstall, *pythonuninstall), chooseStr(*use, *pythonuse))
		return
	}

	if *node {
		if !*list && *install == "" && *uninstall == "" && *use == "" {
			printCurrentNodeVersion()
			return
		}
		handleNode(*list, *install, *uninstall, *use)
		return
	}

	if *golang {
		if !*list && *install == "" && *uninstall == "" && *use == "" {
			printCurrentGoVersion()
			return
		}
		handleGo(*list, *install, *uninstall, *use)
		return
	}

	if *rust {
		if !*list && *install == "" && *uninstall == "" && *use == "" {
			printCurrentRustVersion()
			return
		}
		handleRust(*list, *install, *uninstall, *use)
		return
	}

	// 默认帮助
	fmt.Println("kit 使用示例：")
	fmt.Println("  kit -all                         列出支持的语言")
	fmt.Println("  kit -java                         输出当前 Java 版本")
	fmt.Println("  kit -java -list                   列出支持的 Java 主版本")
	fmt.Println("  kit -java -install 17             安装 Temurin JDK 17")
	fmt.Println("  kit -java -use 17                 切换到 JDK 17（生成 shim）")
	fmt.Println("  kit -java -uninstall 17           卸载 JDK 17")
	fmt.Println("  kit -python                       输出当前 Python 版本")
	fmt.Println("  kit -python -list                 列出支持的 Python 主版本")
	fmt.Println("  kit -python -install 3.12         安装 Python 3.12")
	fmt.Println("  kit -python -use 3.12             切换到 Python 3.12（生成 shim）")
	fmt.Println("  kit -python -uninstall 3.12       卸载 Python 3.12")
	fmt.Println("  kit -node                         输出当前 Node.js 版本")
	fmt.Println("  kit -node -list                   列出支持的 Node.js 版本（LTS）")
	fmt.Println("  kit -node -install 20             安装 Node.js 20（建议配合 nvm-windows）")
	fmt.Println("  kit -node -use 20                 切换到 Node.js 20（生成 shim 或调用 nvm）")
	fmt.Println("  kit -node -uninstall 20           卸载 Node.js 20（如使用 nvm 由 nvm 管理）")
	fmt.Println("  kit -go                           输出当前 Go 版本")
	fmt.Println("  kit -go -list                     列出支持的 Go 版本")
	fmt.Println("  kit -go -install 1.25             安装 Go 1.25")
	fmt.Println("  kit -go -use 1.25                 切换到 Go 1.25（生成 shim）")
	fmt.Println("  kit -go -uninstall 1.25           卸载 Go 1.25")
	fmt.Println("  kit -rust                         输出当前 Rust 版本")
	fmt.Println("  kit -rust -list                   列出支持的 Rust 渠道：stable/beta/nightly")
	fmt.Println("  kit -rust -install stable         安装 rustup 与 stable 工具链")
	fmt.Println("  kit -rust -use stable             切换到 stable（生成 shim 或提示 rustup use）")
	fmt.Println("  kit -rust -uninstall stable       卸载工具链（建议通过 rustup 管理）")
	fmt.Println("  kit -doctor                       自检 PATH、shim、包管理器与当前版本")
}

func listSupportedLangs() {
	fmt.Println("支持的语言：Java, Python, Node.js, Go, Rust")
}

var temurinIDs = map[string]string{
	"8":  "EclipseAdoptium.Temurin.8.JDK",
	"11": "EclipseAdoptium.Temurin.11.JDK",
	"17": "EclipseAdoptium.Temurin.17.JDK",
	"21": "EclipseAdoptium.Temurin.21.JDK",
}

func handleJava(list bool, installVer string, uninstallVer string, useVer string) {
	if runtime.GOOS != "windows" {
		fmt.Println("当前实现主要面向 Windows；其他平台将后续支持。")
	}
	if list {
		fmt.Println("支持的 Java 主版本：8, 11, 17, 21")
	}
	if installVer != "" {
		if cur, _, err0 := detectJavaVersion(); err0 == nil && cur == installVer {
			fmt.Println("当前已是 JDK", installVer, "，无需安装/升级")
		} else if err := wingetInstall(installVer); err != nil {
			fmt.Println("安装失败：", err)
		} else {
			fmt.Println("安装成功：JDK", installVer)
		}
	}
	if uninstallVer != "" {
		if err := wingetUninstall(uninstallVer); err != nil {
			fmt.Println("卸载失败：", err)
		} else {
			fmt.Println("卸载成功：JDK", uninstallVer)
		}
	}
	if useVer != "" {
		jdkRoot, err := findJDKRoot(useVer)
		if err != nil {
			fmt.Println("未找到已安装 JDK", useVer, "。请先执行 'kit -java -install", useVer, "' 或使用系统包管理器安装。")
			return
		}
		if err := ensureShim(jdkRoot); err != nil {
			fmt.Println("生成 shim 失败：", err)
			return
		}
		if err := ensureShimsInPath(); err != nil {
			fmt.Println("加入 PATH 失败（请手动将 %USERPROFILE%/.kit/shims 加入 PATH）：", err)
		} else {
			fmt.Println("已切换到 JDK", useVer, "（通过 shim 生效）。如当前终端未生效，请新开终端。")
		}
	}
}

var pythonIDs = map[string]string{
	"2.7":  "Python.Python.2",
	"3.8":  "Python.Python.3.8",
	"3.9":  "Python.Python.3.9",
	"3.10": "Python.Python.3.10",
	"3.11": "Python.Python.3.11",
	"3.12": "Python.Python.3.12",
	"3.13": "Python.Python.3.13",
	"3.14": "Python.Python.3.14",
}

func handlePython(list bool, installVer string, uninstallVer string, useVer string) {
	if list {
		fmt.Println("支持的 Python 主版本：2.7, 3.8, 3.9, 3.10, 3.11, 3.12, 3.13, 3.14")
	}
	if installVer != "" {
		if cur, _, err := detectPythonVersion(); err == nil && cur == installVer {
			fmt.Println("当前已是 Python", installVer, "，无需安装/升级")
		} else {
			if err := wingetInstallPython(installVer); err != nil {
				fmt.Println("安装失败：", err)
			} else {
				fmt.Println("安装成功：Python", installVer)
			}
		}
	}
	if uninstallVer != "" {
		if err := wingetUninstallPython(uninstallVer); err != nil {
			fmt.Println("卸载失败：", err)
		} else {
			fmt.Println("卸载成功：Python", uninstallVer)
		}
	}
	if useVer != "" {
		root, err := findPythonRoot(useVer)
		if err != nil {
			fmt.Println("未找到已安装 Python", useVer, "。请先执行 'kit -python -install", useVer, "' 或使用系统包管理器安装。")
			return
		}
		if err := ensurePythonShim(root); err != nil {
			fmt.Println("生成 shim 失败：", err)
			return
		}
		if err := ensureShimsInPath(); err != nil {
			fmt.Println("加入 PATH 失败（请手动将 %USERPROFILE%/.kit/shims 加入 PATH）：", err)
		} else {
			fmt.Println("已切换到 Python", useVer, "（通过 shim 生效）。如当前终端未生效，请新开终端。")
		}
	}
}

func handleNode(list bool, installVer string, uninstallVer string, useVer string) {
	if list {
		fmt.Println("支持的 Node.js 主版本（LTS）：18, 20, 22；建议使用 nvm-windows 管理次版本")
	}
	if installVer != "" {
		if cur, _, err := detectNodeVersion(); err == nil && cur == installVer {
			fmt.Println("当前已是 Node.js", installVer, "，无需安装/升级")
		} else {
			if err := wingetInstallNode(installVer); err != nil {
				fmt.Println("安装失败：", err)
			} else {
				fmt.Println("安装成功：Node.js", installVer)
			}
		}
	}
	if uninstallVer != "" {
		if err := wingetUninstallNode(uninstallVer); err != nil {
			fmt.Println("卸载失败：", err)
		} else {
			fmt.Println("卸载成功：Node.js", uninstallVer)
		}
	}
	if useVer != "" {
		if nvm := findNvmWindows(); nvm != "" {
			fmt.Println("检测到 nvm-windows：请执行 'nvm use", useVer, "' 切换版本。")
		}
		root, err := findNodeRoot()
		if err != nil {
			fmt.Println("未找到已安装 Node.js", useVer, "。请先执行 'kit -node -install", useVer, "' 或使用 nvm 安装。")
			return
		}
		if err := ensureNodeShim(root); err != nil {
			fmt.Println("生成 shim 失败：", err)
			return
		}
		if err := ensureShimsInPath(); err != nil {
			fmt.Println("加入 PATH 失败（请手动将 %USERPROFILE%/.kit/shims 加入 PATH）：", err)
		} else {
			fmt.Println("已切换到 Node.js", useVer, "（通过 shim 生效）。如当前终端未生效，请新开终端。")
		}
	}
}

func handleGo(list bool, installVer string, uninstallVer string, useVer string) {
	if list {
		fmt.Println("支持的 Go 主版本：1.20, 1.21, 1.22, 1.25；官方安装包仅支持全局单版本")
	}
	if installVer != "" {
		if cur, _, err := detectGoVersion(); err == nil && strings.HasPrefix(cur, installVer) {
			fmt.Println("当前已是 Go", installVer, "，无需安装/升级")
		} else {
			if err := wingetInstallGo(installVer); err != nil {
				fmt.Println("安装失败：", err)
			} else {
				fmt.Println("安装成功：Go", installVer)
			}
		}
	}
	if uninstallVer != "" {
		if err := wingetUninstallGo(uninstallVer); err != nil {
			fmt.Println("卸载失败：", err)
		} else {
			fmt.Println("卸载成功：Go", uninstallVer)
		}
	}
	if useVer != "" {
		root, err := findGoRoot()
		if err != nil {
			fmt.Println("无法定位 Go 安装路径：", err)
			return
		}
		if err := ensureGoShim(root); err != nil {
			fmt.Println("生成 shim 失败：", err)
			return
		}
		if err := ensureShimsInPath(); err != nil {
			fmt.Println("加入 PATH 失败（请手动将 %USERPROFILE%/.kit/shims 加入 PATH）：", err)
		} else {
			fmt.Println("已生成 Go shim。若需多版本切换，请先安装多个目录或使用社区管理器。")
		}
	}
}

func handleRust(list bool, installVer string, uninstallVer string, useVer string) {
	if list {
		fmt.Println("支持的 Rust 渠道：stable, beta, nightly；推荐通过 rustup 管理具体版本")
	}
	if installVer != "" {
		if _, _, err := detectRustVersion(); err == nil {
			fmt.Println("已检测到 Rust 环境。若需切换版本请使用 rustup。")
		} else {
			if err := wingetInstallRustup(); err != nil {
				fmt.Println("安装 rustup 失败：", err)
				fmt.Println("可选手动安装：")
				fmt.Println("  1) 直接运行：", "https://static.rust-lang.org/rustup/rustup-init.exe")
				fmt.Println("  2) PowerShell：Invoke-WebRequest -Uri https://static.rust-lang.org/rustup/rustup-init.exe -OutFile $env:TEMP\\rustup-init.exe")
				fmt.Println("     然后执行：$env:TEMP\\rustup-init.exe -y --profile minimal --default-toolchain stable")
			} else {
				fmt.Println("已安装 rustup。请运行 'rustup toolchain install", installVer, "' 以安装指定渠道/版本。")
			}
		}
	}
	if uninstallVer != "" {
		fmt.Println("卸载建议通过 'rustup toolchain uninstall", uninstallVer, "' 完成；如需卸载 rustup 可运行 Kit 的卸载命令")
	}
	if useVer != "" {
		root, err := findRustToolchainRoot(useVer)
		if err != nil {
			fmt.Println("未找到 Rust 工具链", useVer, "。请先通过 rustup 安装：'rustup toolchain install", useVer, "'")
			return
		}
		if err := ensureRustShim(root); err != nil {
			fmt.Println("生成 shim 失败：", err)
			return
		}
		if err := ensureShimsInPath(); err != nil {
			fmt.Println("加入 PATH 失败（请手动将 %USERPROFILE%/.kit/shims 加入 PATH）：", err)
		} else {
			fmt.Println("已切换到 Rust", useVer, "（通过 shim 生效）。如当前终端未生效，请新开终端。")
		}
	}
}

func wingetInstallPython(ver string) error {
	id, ok := pythonIDs[ver]
	if !ok {
		return fmt.Errorf("不支持的版本：%s", ver)
	}
	cmd := exec.Command("winget", "install", "-e", "--id", id, "--source", "winget", "--silent", "--accept-package-agreements", "--accept-source-agreements")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func wingetUninstallPython(ver string) error {
	id, ok := pythonIDs[ver]
	if !ok {
		return fmt.Errorf("不支持的版本：%s", ver)
	}
	cmd := exec.Command("winget", "uninstall", "-e", "--id", id, "--source", "winget", "--silent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findPythonRoot(ver string) (string, error) {
	userDir, _ := os.UserHomeDir()
	bases := []string{
		filepath.Join(userDir, "AppData", "Local", "Programs", "Python"),
		`C:\\Program Files`,
		`C:\\Program Files (x86)`,
	}
	token := strings.ReplaceAll(ver, ".", "")
	for _, base := range bases {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			nameLower := strings.ToLower(e.Name())
			if strings.Contains(nameLower, "python"+token) || strings.Contains(nameLower, "python"+strings.ReplaceAll(ver, ".", ".")) || strings.Contains(nameLower, ver) {
				root := filepath.Join(base, e.Name())
				exe := filepath.Join(root, "python.exe")
				if _, err := os.Stat(exe); err == nil {
					return root, nil
				}
				exe2 := filepath.Join(root, "Scripts", "python.exe")
				if _, err := os.Stat(exe2); err == nil {
					return root, nil
				}
			}
		}
	}
	return "", fmt.Errorf("未在常见目录找到 Python %s", ver)
}

func ensurePythonShim(root string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	shimsDir := filepath.Join(userDir, ".kit", "shims")
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return err
	}
	shimPath := filepath.Join(shimsDir, "python.cmd")
	exe := filepath.Join(root, "python.exe")
	if _, err := os.Stat(exe); err != nil {
		exe = filepath.Join(root, "Scripts", "python.exe")
	}
	content := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", exe)
	return os.WriteFile(shimPath, []byte(content), 0644)
}

func printCurrentPythonVersion() {
	ver, full, err := detectPythonVersion()
	if err != nil {
		fmt.Println("未检测到已配置的 Python。可使用：")
		fmt.Println("  kit -python -install 3.12")
	} else {
		fmt.Println("当前 Python 版本：", full)
		if ver != "" {
			fmt.Println("主版本：", ver)
		}
	}
	fmt.Println("可选 Python 主版本：2.7, 3.8, 3.9, 3.10, 3.11, 3.12, 3.13, 3.14")
}

func detectPythonVersion() (string, string, error) {
	if full, err := runPythonVersion("python"); err == nil {
		return parsePythonVersionFromOutput(full)
	}
	if shim := shimPythonPath(); shim != "" {
		if full, err := runPythonVersion(shim); err == nil {
			return parsePythonVersionFromOutput(full)
		}
	}
	candidates := findPythonExecutables()
	for _, exe := range candidates {
		if full, err := runPythonVersion(exe); err == nil {
			return parsePythonVersionFromOutput(full)
		}
	}
	return "", "", fmt.Errorf("未找到 python 可执行文件")
}

func runPythonVersion(pyCmd string) (string, error) {
	cmd := exec.Command(pyCmd, "--version")
	outStderr, _ := cmd.StderrPipe()
	outStdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return "", err
	}
	b := &strings.Builder{}
	s1 := bufio.NewScanner(outStdout)
	for s1.Scan() {
		b.WriteString(s1.Text())
		b.WriteString("\n")
	}
	s2 := bufio.NewScanner(outStderr)
	for s2.Scan() {
		b.WriteString(s2.Text())
		b.WriteString("\n")
	}
	_ = cmd.Wait()
	return b.String(), nil
}

func parsePythonVersionFromOutput(out string) (string, string, error) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return "", "", fmt.Errorf("无版本输出")
	}
	full := lines[0]
	lower := strings.ToLower(full)
	ver := ""
	if strings.HasPrefix(lower, "python ") {
		verStr := strings.TrimSpace(strings.TrimPrefix(full, "Python "))
		parts := strings.Split(verStr, ".")
		if len(parts) >= 2 {
			ver = parts[0] + "." + parts[1]
		} else if len(parts) == 1 {
			ver = parts[0]
		}
	}
	return ver, full, nil
}

func shimPythonPath() string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(userDir, ".kit", "shims", "python.cmd")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

func findPythonExecutables() []string {
	userDir, _ := os.UserHomeDir()
	bases := []string{
		filepath.Join(userDir, "AppData", "Local", "Programs", "Python"),
		`C:\\Program Files`,
		`C:\\Program Files (x86)`,
	}
	var res []string
	for _, base := range bases {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			root := filepath.Join(base, e.Name())
			exe := filepath.Join(root, "python.exe")
			if _, err := os.Stat(exe); err == nil {
				res = append(res, exe)
			}
			exe2 := filepath.Join(root, "Scripts", "python.exe")
			if _, err := os.Stat(exe2); err == nil {
				res = append(res, exe2)
			}
		}
	}
	return res
}

func chooseStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func wingetInstall(ver string) error {
	id, ok := temurinIDs[ver]
	if !ok {
		return fmt.Errorf("不支持的版本：%s", ver)
	}
	cmd := exec.Command("winget", "install", "-e", "--id", id, "--source", "winget", "--silent", "--accept-package-agreements", "--accept-source-agreements")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func wingetUninstall(ver string) error {
	id, ok := temurinIDs[ver]
	if !ok {
		return fmt.Errorf("不支持的版本：%s", ver)
	}
	cmd := exec.Command("winget", "uninstall", "-e", "--id", id, "--source", "winget", "--silent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func wingetInstallNode(ver string) error {
	id, ok := nodeIDs[strings.ToLower(ver)]
	if !ok {
		id = nodeIDs["lts"]
	}
	args := []string{"install", "-e", "--id", id, "--source", "winget", "--silent", "--accept-package-agreements", "--accept-source-agreements"}
	// 如用户提供完整版本（含点），尝试传递 --version
	if strings.Contains(ver, ".") {
		args = append(args, "--version", ver)
	}
	cmd := exec.Command("winget", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func wingetUninstallNode(ver string) error {
	id, ok := nodeIDs[strings.ToLower(ver)]
	if !ok {
		id = nodeIDs["lts"]
	}
	cmd := exec.Command("winget", "uninstall", "-e", "--id", id, "--source", "winget", "--silent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func wingetInstallGo(ver string) error {
	id, ok := goIDs[ver]
	if !ok {
		id = "GoLang.Go"
	}
	args := []string{"install", "-e", "--id", id, "--source", "winget", "--silent", "--accept-package-agreements", "--accept-source-agreements"}
	if strings.Contains(ver, ".") {
		args = append(args, "--version", ver)
	}
	cmd := exec.Command("winget", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func wingetUninstallGo(ver string) error {
	id, ok := goIDs[ver]
	if !ok {
		id = "GoLang.Go"
	}
	cmd := exec.Command("winget", "uninstall", "-e", "--id", id, "--source", "winget", "--silent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func wingetInstallRustup() error {
	id := rustIDs["rustup"]
	cmd := exec.Command("winget", "install", "-e", "--id", id, "--source", "winget", "--silent", "--accept-package-agreements", "--accept-source-agreements")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return downloadAndInstallRustup()
	}
	return nil
}

func downloadAndInstallRustup() error {
	url := "https://static.rust-lang.org/rustup/rustup-init.exe"
	tmp := filepath.Join(os.TempDir(), "rustup-init.exe")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	cmd := exec.Command(tmp, "-y", "--profile", "minimal", "--default-toolchain", "stable")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findJDKRoot(ver string) (string, error) {
	baseDirs := []string{
		`C:\\Program Files\\Eclipse Adoptium`,
		`C:\\Program Files\\Java`,
		`C:\\Program Files (x86)\\Eclipse Adoptium`,
		`C:\\Program Files (x86)\\Java`,
	}
	for _, base := range baseDirs {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			nameLower := strings.ToLower(e.Name())
			if strings.Contains(nameLower, "jdk"+ver) || strings.Contains(nameLower, "jdk-"+ver) || strings.Contains(nameLower, "jdk"+ver+"u") {
				root := filepath.Join(base, e.Name())
				javaBin := filepath.Join(root, "bin", "java.exe")
				if _, err := os.Stat(javaBin); err == nil {
					return root, nil
				}
			}
		}
	}
	return "", fmt.Errorf("未在常见目录找到 JDK %s", ver)
}

func ensureShim(jdkRoot string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	shimsDir := filepath.Join(userDir, ".kit", "shims")
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return err
	}
	shimPath := filepath.Join(shimsDir, "java.cmd")
	content := fmt.Sprintf("@echo off\r\nset \"JAVA_HOME=%s\"\r\n\"%s\\bin\\java.exe\" %%*\r\n", jdkRoot, jdkRoot)
	return os.WriteFile(shimPath, []byte(content), 0644)
}

func ensureNodeShim(root string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	shimsDir := filepath.Join(userDir, ".kit", "shims")
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return err
	}
	shimPath := filepath.Join(shimsDir, "node.cmd")
	exe := filepath.Join(root, "node.exe")
	content := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", exe)
	return os.WriteFile(shimPath, []byte(content), 0644)
}

func ensureGoShim(root string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	shimsDir := filepath.Join(userDir, ".kit", "shims")
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return err
	}
	shimPath := filepath.Join(shimsDir, "go.cmd")
	exe := filepath.Join(root, "bin", "go.exe")
	content := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", exe)
	return os.WriteFile(shimPath, []byte(content), 0644)
}

func ensureRustShim(toolchainBin string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	shimsDir := filepath.Join(userDir, ".kit", "shims")
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return err
	}
	cargoShim := filepath.Join(shimsDir, "cargo.cmd")
	rustcShim := filepath.Join(shimsDir, "rustc.cmd")
	cargoExe := filepath.Join(toolchainBin, "cargo.exe")
	rustcExe := filepath.Join(toolchainBin, "rustc.exe")
	contentCargo := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", cargoExe)
	contentRustc := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", rustcExe)
	if err := os.WriteFile(cargoShim, []byte(contentCargo), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(rustcShim, []byte(contentRustc), 0644); err != nil {
		return err
	}
	return nil
}

func ensureShimsInPath() error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	shimsDir := filepath.Join(userDir, ".kit", "shims")
	// 检查当前用户 PATH 是否已包含
	currentPath := os.Getenv("PATH")
	if strings.Contains(strings.ToLower(currentPath), strings.ToLower(shimsDir)) {
		return nil
	}
	// 使用 setx 追加（注意：不会影响当前进程，需要新开终端）
	newPath := currentPath + ";" + shimsDir
	cmd := exec.Command("cmd", "/C", "setx", "PATH", newPath)
	// 为了避免输出过长，读最后一行
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	_ = cmd.Start()
	_ = consume(stdout)
	_ = consume(stderr)
	return cmd.Wait()
}

func consume(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		// 丢弃输出
	}
	return scanner.Err()
}

func printCurrentJavaVersion() {
	ver, full, err := detectJavaVersion()
	if err != nil {
		fmt.Println("未检测到已配置的 Java。可使用：")
		fmt.Println("  kit -java -install 17")
	} else {
		fmt.Println("当前 Java 版本：", full)
		if ver != "" {
			fmt.Println("主版本：", ver)
		}
	}
	fmt.Println("可选 Java 主版本：8, 11, 17, 21")
}

func detectJavaVersion() (string, string, error) {
	// 1) 直接调用 PATH 中的 java
	if full, err := runJavaVersion("java"); err == nil {
		return parseJavaVersionFromOutput(full)
	}
	// 2) 尝试使用 shim
	if shim := shimJavaPath(); shim != "" {
		if full, err := runJavaVersion(shim); err == nil {
			return parseJavaVersionFromOutput(full)
		}
	}
	// 3) 在常见目录里查找并尝试执行
	candidates := findJavaExecutables()
	for _, exe := range candidates {
		if full, err := runJavaVersion(exe); err == nil {
			return parseJavaVersionFromOutput(full)
		}
	}
	return "", "", fmt.Errorf("未找到 java 可执行文件")
}

func printCurrentNodeVersion() {
	ver, full, err := detectNodeVersion()
	if err != nil {
		fmt.Println("未检测到已配置的 Node.js。可使用：")
		fmt.Println("  kit -node -install 20")
	} else {
		fmt.Println("当前 Node.js 版本：", full)
		if ver != "" {
			fmt.Println("主版本：", ver)
		}
	}
	fmt.Println("可选 Node.js 主版本：18, 20, 22（LTS）")
}

func detectNodeVersion() (string, string, error) {
	if full, err := runCmdVersion("node", "--version"); err == nil {
		return parseNodeVersionFromOutput(full)
	}
	if shim := shimNodePath(); shim != "" {
		if full, err := runCmdVersion(shim, "--version"); err == nil {
			return parseNodeVersionFromOutput(full)
		}
	}
	exe := filepath.Join(`C:\\Program Files\\nodejs`, "node.exe")
	if _, err := os.Stat(exe); err == nil {
		if full, err := runCmdVersion(exe, "--version"); err == nil {
			return parseNodeVersionFromOutput(full)
		}
	}
	return "", "", fmt.Errorf("未找到 node 可执行文件")
}

func parseNodeVersionFromOutput(out string) (string, string, error) {
	line := strings.TrimSpace(out)
	if line == "" {
		return "", "", fmt.Errorf("无版本输出")
	}
	full := line
	// 形式：v20.12.1
	if strings.HasPrefix(strings.ToLower(line), "v") {
		verStr := strings.TrimPrefix(line, "v")
		parts := strings.Split(verStr, ".")
		if len(parts) >= 1 {
			return parts[0], full, nil
		}
	}
	return "", full, nil
}

func shimNodePath() string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(userDir, ".kit", "shims", "node.cmd")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

func findNodeRoot() (string, error) {
	bases := []string{
		`C:\\Program Files\\nodejs`,
	}
	for _, b := range bases {
		exe := filepath.Join(b, "node.exe")
		if _, err := os.Stat(exe); err == nil {
			return b, nil
		}
	}
	return "", fmt.Errorf("未在常见目录找到 Node.js")
}

func findNvmWindows() string {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		return ""
	}
	nvm := filepath.Join(appdata, "nvm", "nvm.exe")
	if _, err := os.Stat(nvm); err == nil {
		return nvm
	}
	return ""
}

func printCurrentGoVersion() {
	ver, full, err := detectGoVersion()
	if err != nil {
		fmt.Println("未检测到已配置的 Go。可使用：")
		fmt.Println("  kit -go -install 1.25")
	} else {
		fmt.Println("当前 Go 版本：", full)
		if ver != "" {
			fmt.Println("主版本：", ver)
		}
	}
	fmt.Println("可选 Go 主版本：1.20, 1.21, 1.22, 1.25")
}

func detectGoVersion() (string, string, error) {
	if full, err := runCmdVersion("go", "version"); err == nil {
		return parseGoVersionFromOutput(full)
	}
	if shim := shimGoPath(); shim != "" {
		if full, err := runCmdVersion(shim, "version"); err == nil {
			return parseGoVersionFromOutput(full)
		}
	}
	exe := filepath.Join(`C:\\Program Files\\Go\\bin`, "go.exe")
	if _, err := os.Stat(exe); err == nil {
		if full, err := runCmdVersion(exe, "version"); err == nil {
			return parseGoVersionFromOutput(full)
		}
	}
	return "", "", fmt.Errorf("未找到 go 可执行文件")
}

func parseGoVersionFromOutput(out string) (string, string, error) {
	line := strings.TrimSpace(out)
	if line == "" {
		return "", "", fmt.Errorf("无版本输出")
	}
	full := line
	// 形式：go version go1.25.5 windows/amd64
	lower := strings.ToLower(line)
	idx := strings.Index(lower, "go1")
	if idx >= 0 {
		verStr := lower[idx+2:]
		parts := strings.Split(verStr, " ")
		verPart := parts[0] // 1.25.5
		segs := strings.Split(verPart, ".")
		if len(segs) >= 2 {
			return segs[0] + "." + segs[1], full, nil
		}
		return segs[0], full, nil
	}
	return "", full, nil
}

func shimGoPath() string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(userDir, ".kit", "shims", "go.cmd")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

func findGoRoot() (string, error) {
	base := `C:\\Program Files\\Go`
	exe := filepath.Join(base, "bin", "go.exe")
	if _, err := os.Stat(exe); err == nil {
		return base, nil
	}
	return "", fmt.Errorf("未在常见目录找到 Go")
}

func printCurrentRustVersion() {
	_, full, err := detectRustVersion()
	if err != nil {
		fmt.Println("未检测到已配置的 Rust。可使用：")
		fmt.Println("  kit -rust -install rustup")
	}
	if err == nil {
		fmt.Println("当前 Rust 版本：", full)
	}
	ch := rustSelectableChannels()
	if len(ch) > 0 {
		fmt.Println("可选 Rust 渠道：", strings.Join(ch, ", "))
	}
}

func detectRustVersion() (string, string, error) {
	if full, err := runCmdVersion("rustc", "--version"); err == nil {
		return parseRustVersionFromOutput(full)
	}
	if shim := shimRustcPath(); shim != "" {
		if full, err := runCmdVersion(shim, "--version"); err == nil {
			return parseRustVersionFromOutput(full)
		}
	}
	// 尝试查找 stable 工具链
	if home, err := os.UserHomeDir(); err == nil {
		tc := filepath.Join(home, ".rustup", "toolchains")
		entries, _ := os.ReadDir(tc)
		for _, e := range entries {
			bin := filepath.Join(tc, e.Name(), "bin", "rustc.exe")
			if _, err := os.Stat(bin); err == nil {
				if full, err := runCmdVersion(bin, "--version"); err == nil {
					return parseRustVersionFromOutput(full)
				}
			}
		}
	}
	return "", "", fmt.Errorf("未找到 rustc 可执行文件")
}

func parseRustVersionFromOutput(out string) (string, string, error) {
	line := strings.TrimSpace(out)
	if line == "" {
		return "", "", fmt.Errorf("无版本输出")
	}
	full := line // 例如：rustc 1.74.1 (..)
	return "", full, nil
}

func shimRustcPath() string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(userDir, ".kit", "shims", "rustc.cmd")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

func findRustToolchainRoot(channel string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	tc := filepath.Join(home, ".rustup", "toolchains")
	entries, err := os.ReadDir(tc)
	if err != nil {
		return "", fmt.Errorf("未找到 rustup toolchains 目录")
	}
	channelLower := strings.ToLower(channel)
	for _, e := range entries {
		name := strings.ToLower(e.Name())
		if strings.Contains(name, channelLower) && strings.Contains(name, "windows") && strings.Contains(name, "msvc") {
			bin := filepath.Join(tc, e.Name(), "bin")
			if _, err := os.Stat(filepath.Join(bin, "rustc.exe")); err == nil {
				return bin, nil
			}
		}
	}
	return "", fmt.Errorf("未找到匹配的工具链：%s", channel)
}

func rustSelectableChannels() []string {
	base := []string{"stable", "beta", "nightly"}
	m := map[string]bool{}
	for _, b := range base {
		m[b] = true
	}
	if home, err := os.UserHomeDir(); err == nil {
		tc := filepath.Join(home, ".rustup", "toolchains")
		entries, _ := os.ReadDir(tc)
		for _, e := range entries {
			name := strings.ToLower(e.Name())
			if i := strings.Index(name, "-"); i > 0 {
				ch := name[:i]
				if ch == "stable" || ch == "beta" || ch == "nightly" {
					m[ch] = true
				}
			}
		}
	}
	var res []string
	for _, k := range []string{"stable", "beta", "nightly"} {
		if m[k] {
			res = append(res, k)
		}
	}
	return res
}

func runJavaVersion(javaCmd string) (string, error) {
	cmd := exec.Command(javaCmd, "-version")
	// Java 版本输出常在 stderr
	outStderr, _ := cmd.StderrPipe()
	outStdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return "", err
	}
	buf := &strings.Builder{}
	copyToBuilder := func(r io.Reader) {
		s := bufio.NewScanner(r)
		for s.Scan() {
			buf.WriteString(s.Text())
			buf.WriteString("\n")
		}
	}
	copyToBuilder(outStdout)
	copyToBuilder(outStderr)
	if err := cmd.Wait(); err != nil {
		// 某些实现返回非 0，但版本信息仍在输出
		if buf.Len() == 0 {
			return "", err
		}
	}
	return buf.String(), nil
}

func runCmdVersion(cmdPath string, arg string) (string, error) {
	cmd := exec.Command(cmdPath, arg)
	outStderr, _ := cmd.StderrPipe()
	outStdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return "", err
	}
	b := &strings.Builder{}
	s1 := bufio.NewScanner(outStdout)
	for s1.Scan() {
		b.WriteString(s1.Text())
		b.WriteString("\n")
	}
	s2 := bufio.NewScanner(outStderr)
	for s2.Scan() {
		b.WriteString(s2.Text())
		b.WriteString("\n")
	}
	_ = cmd.Wait()
	return b.String(), nil
}

func parseJavaVersionFromOutput(out string) (string, string, error) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return "", "", fmt.Errorf("无版本输出")
	}
	first := strings.ToLower(lines[0])
	full := lines[0]
	// 可能的形式：
	// openjdk version "17.0.7" 2023-...
	// java version "1.8.0_372"
	// java version "21" 2024-...
	ver := ""
	if idx := strings.Index(first, "\""); idx >= 0 {
		firstQuote := idx
		rest := first[firstQuote+1:]
		if j := strings.Index(rest, "\""); j >= 0 {
			versionStr := rest[:j]
			// 提取主版本：1.8 -> 8；17.0.7 -> 17；21 -> 21
			parts := strings.Split(versionStr, ".")
			if len(parts) > 0 {
				main := parts[0]
				if main == "1" && len(parts) > 1 {
					ver = parts[1]
				} else {
					ver = main
				}
			}
		}
	}
	return ver, full, nil
}

func shimJavaPath() string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(userDir, ".kit", "shims", "java.cmd")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

func findJavaExecutables() []string {
	baseDirs := []string{
		`C:\\Program Files\\Eclipse Adoptium`,
		`C:\\Program Files\\Java`,
		`C:\\Program Files (x86)\\Eclipse Adoptium`,
		`C:\\Program Files (x86)\\Java`,
	}
	var res []string
	for _, base := range baseDirs {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			root := filepath.Join(base, e.Name())
			javaBin := filepath.Join(root, "bin", "java.exe")
			if _, err := os.Stat(javaBin); err == nil {
				res = append(res, javaBin)
			}
		}
	}
	return res
}

var nodeIDs = map[string]string{
	// 建议使用 nvm-windows 管理多版本
	"nvm": "CoreyButler.NVMforWindows",
	"lts": "OpenJS.NodeJS.LTS",
	"18":  "OpenJS.NodeJS.LTS",
	"20":  "OpenJS.NodeJS.LTS",
	"22":  "OpenJS.NodeJS.LTS",
}

var goIDs = map[string]string{
	// 官方 Go 安装包（只能安装一个版本）
	"1.20": "GoLang.Go",
	"1.21": "GoLang.Go",
	"1.22": "GoLang.Go",
	"1.25": "GoLang.Go",
}

var rustIDs = map[string]string{
	// 安装 rustup（推荐），再由 rustup 管理工具链
	"rustup": "Rustlang.Rustup",
}

func runDoctor() {
	fmt.Println("环境自检：")
	fmt.Println("  包管理器：")
	if _, err := exec.LookPath("winget"); err == nil {
		fmt.Println("   - winget: 可用")
	} else {
		fmt.Println("   - winget: 不可用")
	}
	fmt.Println("  PATH 与 shim：")
	userDir, _ := os.UserHomeDir()
	shimsDir := filepath.Join(userDir, ".kit", "shims")
	p := os.Getenv("PATH")
	if strings.Contains(strings.ToLower(p), strings.ToLower(shimsDir)) {
		fmt.Println("   - PATH: 已包含", shimsDir)
	} else {
		fmt.Println("   - PATH: 未包含", shimsDir)
	}
	checkShim := func(name string) {
		fp := filepath.Join(shimsDir, name)
		if _, err := os.Stat(fp); err == nil {
			fmt.Println("   - ", name, ": 存在")
		} else {
			fmt.Println("   - ", name, ": 缺失")
		}
	}
	checkShim("java.cmd")
	checkShim("python.cmd")
	checkShim("node.cmd")
	checkShim("go.cmd")
	checkShim("cargo.cmd")
	checkShim("rustc.cmd")
	fmt.Println("  当前版本：")
	if _, full, err := detectJavaVersion(); err == nil {
		fmt.Println("   - Java:", full)
	} else {
		fmt.Println("   - Java: 未检测到")
	}
	if _, full, err := detectPythonVersion(); err == nil {
		fmt.Println("   - Python:", full)
	} else {
		fmt.Println("   - Python: 未检测到")
	}
	if _, full, err := detectNodeVersion(); err == nil {
		fmt.Println("   - Node.js:", full)
	} else {
		fmt.Println("   - Node.js: 未检测到")
	}
	if _, full, err := detectGoVersion(); err == nil {
		fmt.Println("   - Go:", full)
	} else {
		fmt.Println("   - Go: 未检测到")
	}
	if _, full, err := detectRustVersion(); err == nil {
		fmt.Println("   - Rust:", full)
	} else {
		fmt.Println("   - Rust: 未检测到")
	}
	fmt.Println("  提示：如需切换版本，先使用 -install 安装目标版本，再使用 -use 生成 shim 并生效")
}
