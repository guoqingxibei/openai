package util

import (
	"github.com/disintegration/imaging"
	"image"
	"os"
	"path/filepath"
	"strconv"
)

func SplitImage(inputImage string) ([]string, error) {
	// 打开输入的图像文件
	src, err := imaging.Open(inputImage)
	if err != nil {
		return nil, err
	}

	// 获取图像的大小
	srcBounds := src.Bounds()
	width := srcBounds.Dx()
	height := srcBounds.Dy()

	// 计算分割后的图像大小
	splitWidth := width / 2
	splitHeight := height / 2

	// 获取输入图像的扩展名/后缀
	extension := filepath.Ext(inputImage)

	// 删除扩展名以便于后面添加序号
	base := inputImage[:len(inputImage)-len(extension)]

	// 创建分割后的图像的名字列表
	var outputImages []string

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			startX := i * splitWidth
			startY := j * splitHeight

			// 分割图像
			part := imaging.Crop(src, image.Rect(startX, startY, startX+splitWidth, startY+splitHeight))

			// 构建新图像的文件名
			outputImage := base + "_" + strconv.Itoa(i*2+j) + extension
			err = imaging.Save(part, outputImage)

			if err != nil {
				return nil, err
			}

			outputImages = append(outputImages, outputImage)
		}
	}

	return outputImages, nil
}

func GetImageSize(imageFile string) (width int, height int, err error) {
	reader, err := os.Open(imageFile)
	if err != nil {
		return
	}

	defer reader.Close()
	im, _, err := image.DecodeConfig(reader)
	if err != nil {
		return
	}

	return im.Width, im.Height, err
}
