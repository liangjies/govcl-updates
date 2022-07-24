package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	_ "github.com/ying32/govcl/pkgs/winappres"
	"github.com/ying32/govcl/vcl"
	"github.com/ying32/govcl/vcl/win"
)

var (
	Server    string
	Port      string
	MainEXE   string
	UpdateEXE string
)

// go build -i -ldflags="-s -w -H windowsgui" -tags tempdll
func main() {
	// 加载配置文件
	iniFile := vcl.NewIniFile(".\\Config.ini")
	defer iniFile.Free()
	// 读取配置文件
	Server = iniFile.ReadString("System", "Server", "")
	Port = iniFile.ReadString("System", "Port", "")
	MainEXE = iniFile.ReadString("System", "MainEXE", "")
	UpdateEXE = iniFile.ReadString("System", "UpdateEXE", "")
	// 检测是否有新版本
	if !checkUpdate() {
		win.MessageBox(0, "当前已是最新版本，无需更新", "提示", win.MB_OK+win.MB_ICONINFORMATION)
		return
	}
	// 启动下载程序
	err := downUpdate()
	if err != nil {
		win.MessageBox(0, "程序更新失败，下载失败", "错误", win.MB_OK+win.MB_ICONERROR)
		return
	}
	bytestr, err := ioutil.ReadFile("update.tmp")
	if err != nil {
		fmt.Println(err)
	}
	// 校验文件
	if !verifyUpdate() {
		win.MessageBox(0, "程序更新失败，校验失败！", "错误", win.MB_OK+win.MB_ICONERROR)
		return
	}

	// ioutil.WriteFile将读取到的文件写入目标文件中
	err2 := ioutil.WriteFile(MainEXE, bytestr, os.ModePerm)
	if err2 != nil {
		fmt.Println(err2)
	} else {
		err := os.Remove("update.tmp")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("拷贝成功。。。")
	}

	win.MessageBox(0, "程序更新成功，请重新打开程序", "程序更新", win.MB_OK+win.MB_ICONINFORMATION)
}

// 下载更新
func downUpdate() error {

	// 定义结果文件名
	fileName := "update.tmp"
	// 下载URL
	updateURL := "http://" + Server + ":" + Port + "/updates/process-order/update.exe"
	// 获取数据
	resp, err := http.Get(updateURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建一个文件用于保存
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	// 然后将响应流和文件流对接起来
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// 校验程序
func verifyUpdate() bool {
	updateURL := "http://" + Server + ":" + Port + "/updates/process-order/md5.txt"
	// 捕获异常
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("检测更新异常:", err)
		}
	}()
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(updateURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	// 检测是否有新版本
	f, err := os.Open("update.tmp")
	if err != nil {
		fmt.Println("Open file error:", err)
		return false
	}
	defer f.Close()
	// 获取文件的MD5值
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		fmt.Println("Copy file error:", err)
		return false
	}
	FileMd5 := h.Sum(nil)
	return fmt.Sprintf("%x", FileMd5) == string(body)
}

// 检测程序是否有更新
func checkUpdate() bool {
	updateURL := "http://" + Server + ":" + Port + "/updates/process-order/md5.txt"
	// 捕获异常
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("检测更新异常:", err)
		}
	}()
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(updateURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	// 检测是否有新版本
	f, err := os.Open(MainEXE)
	if err != nil {
		fmt.Println("Open file error:", err)
		return false
	}
	defer f.Close()
	// 获取文件的MD5值
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		fmt.Println("Copy file error:", err)
		return false
	}
	FileMd5 := h.Sum(nil)
	fmt.Println(fmt.Sprintf("FileMd5:%x", FileMd5))
	fmt.Println("body:", string(body))
	return fmt.Sprintf("%x", FileMd5) != string(body)
}
