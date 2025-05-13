package services

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/Petryanin/love-bot/internal/clients"
	"github.com/fogleman/gg"
)

type ImageComplimentService struct {
	Client    *clients.CatAAS
	FontPath  string
	FontSize  float64
	ImgWidth  int
	ImgHeight int
}

// конструктор
func NewImageComplimentService(client *clients.CatAAS, fontPath string) *ImageComplimentService {
	return &ImageComplimentService{
		Client:    client,
		FontPath:  fontPath,
		FontSize:  45,
		ImgWidth:  700,
		ImgHeight: 700,
	}
}

// Generate возвращает PNG-байты картинки с наклеенным комплиментом
func (s *ImageComplimentService) Generate(compliment string) ([]byte, error) {
	// 1. Получаем чистое фото кота нужного размера
	baseImg, err := s.Client.Image(s.ImgHeight, s.ImgWidth)
	if err != nil {
		return nil, err
	}

	// 2. Создаём контекст для рисования
	dc := gg.NewContext(s.ImgWidth, s.ImgHeight)
	dc.DrawImage(*baseImg, 0, 0)

	// 3. Загружаем шрифт
	if err := dc.LoadFontFace(s.FontPath, s.FontSize); err != nil {
		return nil, fmt.Errorf("failed to load font %s: %w", s.FontPath, err)
	}

	// 4. Готовим текст: разбиваем на строки по ширине
	wrappedText := dc.WordWrap(compliment, float64(s.ImgWidth)*0.9)

	// параметры тени
	shadowOffsetX := 3.0                    // смещение вправо
	shadowOffsetY := 3.0                    // смещение вниз
	shadowColor := color.RGBA{0, 0, 0, 220} // почти чёрная с прозрачностью

	// 6. Рисуем сам текст белым по центру
	x := s.calculateHorizontalOffset()
	y := s.calculateVerticalOffset(dc, wrappedText)
	for _, line := range wrappedText {
		// 1) рисуем тень
		dc.SetColor(shadowColor)
		dc.DrawStringAnchored(line, x+shadowOffsetX, y+shadowOffsetY, 0.5, 0.5)

		// 2) рисуем основной текст поверх
		dc.SetColor(color.White)
		dc.DrawStringAnchored(line, x, y, 0.5, 0.5)
		y += dc.FontHeight() + 10
	}

	// 7. Кодируем в PNG
	buf := new(bytes.Buffer)
	if err := dc.EncodePNG(buf); err != nil {
		return nil, fmt.Errorf("failed to encode png from %v: %w", buf, err)
	}
	return buf.Bytes(), nil
}

func (s *ImageComplimentService) calculateHorizontalOffset() float64 {
	return float64(s.ImgWidth) / 2
}

func (s *ImageComplimentService) calculateVerticalOffset(dc *gg.Context, wrappedText []string) float64 {
	textRows := len(wrappedText)
	multiplier := 3.0

	if textRows > 4 {
		multiplier = 1.3
	} else if textRows > 2 {
		multiplier = 1.5
	} else if textRows > 1 {
		multiplier = 2
	}

	return float64(s.ImgHeight) - dc.FontHeight()*float64(textRows)*multiplier
}
