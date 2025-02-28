package utils

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/image/draw"
)

// 将图片宽度压缩为指定宽度newWidth
// forceResize 是否强制转换大小, 例如如果图片大小不足newWidth可以执行放大
func ComposeImg(imgPath, outputPath string, newWidth int, forceResize bool) error {
	// targetWidth := 1000
	file, err := os.Open(imgPath)
	if err != nil {
		return errors.New("打开文件失败:" + err.Error())
	}
	defer file.Close()
	img, formatName, err := image.Decode(file)
	if err != nil {
		return errors.New("解码文件失败:" + err.Error())
	}
	formatName = strings.ToLower(formatName)
	fmt.Println("文件格式:", formatName)
	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return errors.New("创建输出文件失败:" + err.Error())
	}
	defer outFile.Close()

	// 计算宽高
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// 计算缩放了多少
	ratio := float64(width) / float64(newWidth)
	// 缩放比和新高度
	newHeight := float64(height) / ratio
	log.Printf("原宽度%v,原高度:%v, 宽高比:%v, 新宽度:%v，新高度%v", width, height, ratio, newWidth, newHeight)

	// 如果大于newWidth则执行压缩
	if width > newWidth {
		// 创建新的图片
		newImg := image.NewRGBA(image.Rect(0, 0, newWidth, int(newHeight)))
		draw.CatmullRom.Scale(newImg, newImg.Bounds(), img, bounds, draw.Over, nil)
		// 编码并保存压缩后的图片
		if formatName == "jpg" || formatName == "jpeg" {
			jpeg.Encode(outFile, newImg, &jpeg.Options{Quality: 90})
		} else {
			png.Encode(outFile, newImg)
		}

	} else {
		if forceResize {
			fmt.Println("宽度不大于目标宽度:", bounds, "执行放大图片操作")
			newImg := image.NewRGBA(image.Rect(0, 0, newWidth, int(newHeight)))
			draw.CatmullRom.Scale(newImg, newImg.Bounds(), img, bounds, draw.Over, nil)
			jpeg.Encode(outFile, newImg, &jpeg.Options{Quality: 90})
		} else {
			// 保持原大小
			// newImg := image.NewRGBA(image.Rect(0, 0, width, height))
			// draw.CatmullRom.Scale(newImg, newImg.Bounds(), img, bounds, draw.Over, nil)
			// jpeg.Encode(outFile, newImg, &jpeg.Options{Quality: 90})

			// 直接复制一份,不用浪费时间去编码
			file.Seek(0, io.SeekStart) // 将文件指针重置
			written, err := io.Copy(outFile, file)
			if err != nil {
				log.Println("复制文件错误:", err)
			} else {
				log.Println("直接复制文件完成", written)
			}
		}
	}

	return nil
}

// 返回图片宽度, 如果读取失败则返回width为0和error信息
func ImageWidth(imgPath string) (int, error) {
	file, err := os.Open(imgPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return 0, err
	}
	return img.Bounds().Max.X, nil
}
