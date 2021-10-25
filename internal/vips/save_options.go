package vips

type JpegSaveOptions struct {
	Quality            int
	Interlace          bool
	OptimizeCoding     bool
	SubsampleMode      SubsampleMode
	TrellisQuant       bool
	OvershootDeringing bool
	OptimizeScans      bool
	QuantTable         int
}

type PngSaveOptions struct {
	Compression int
	Interlace   bool
	Quantize    bool
	Colors      int
}

type WebpSaveOptions struct {
	Quality         int
	Lossless        bool
	SmartSubsample  bool
	ReductionEffort int
}

type AvifSaveOptions struct {
	Quality  int
	Lossless bool
	Speed    int
}

// MozJpeg default save options
func NewJpegSaveOptions() (options JpegSaveOptions) {
	options.Quality = 75
	options.Interlace = true
	options.OptimizeCoding = true
	options.SubsampleMode = VipsForeignSubsampleAuto
	options.TrellisQuant = true
	options.OvershootDeringing = true
	options.OptimizeScans = true
	options.QuantTable = 3
	return
}

func NewPngSaveOptions() (options PngSaveOptions) {
	options.Compression = 6
	options.Interlace = false
	options.Quantize = false
	options.Colors = 256
	return
}

func NewWebpSaveOptions() (options WebpSaveOptions) {
	options.Quality = 75
	options.Lossless = false
	options.SmartSubsample = true
	options.ReductionEffort = 4
	return
}

func NewAvifSaveptions() (options AvifSaveOptions) {
	options.Quality = 80
	options.Lossless = false
	options.Speed = 5
	return
}
