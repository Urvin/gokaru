package thumbnailer

import (
	"errors"
	"fmt"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/fileinfo"
	"github.com/urvin/gokaru/internal/helper"
	"github.com/urvin/gokaru/internal/logging"
	"github.com/urvin/gokaru/internal/vips"
	"io/ioutil"
	"math"
	"os"
	"strconv"
)

type thumbnailer struct {
	logger  logging.Logger
	imageId uint64
}

func (t *thumbnailer) Thumbnail(origin []byte, options ThumbnailOptions) (thumbnail []byte, later func([]byte) ([]byte, error), err error) {
	originMime := fileinfo.MimeByData(origin)
	originType := vips.ImageTypeByByMime(originMime)
	if originType == vips.ImageTypeUnknown {
		err = errors.New("unknown origin image type")
		return
	}

	if options.ImageType() == vips.ImageTypeUnknown {
		err = errors.New("unknown destination image type")
		return
	}

	animationSupport := originType.SupportsAnimation() && options.ImageType().SupportsAnimation()
	pages := 1
	if animationSupport {
		pages = -1
	}

	image := new(vips.Image)
	defer image.Clear()
	defer vips.Cleanup()

	err = image.Load(origin, originType, 1, 1.0, pages)
	if err != nil {
		return
	}

	imageId := t.newImageId()
	t.logger.Info(fmt.Sprintf("[thumbnailer] #%d NEW THUMBNAIL", imageId))

	if animationSupport && image.IsAnimated() {
		err = t.transformFrames(imageId, origin, image, originType, &options)
	} else {
		err = t.transformFrame(imageId, image, &options)
	}
	if err != nil {
		return
	}

	err = image.CopyMemory()
	if err != nil {
		return
	}

	q := t.getQuality(uint(image.Width()), uint(image.Height()), options.ImageType())

	switch options.ImageType() {
	case vips.ImageTypeJPEG:
		jo := vips.NewJpegSaveOptions()
		jo.Quality = int(q.Quality)
		thumbnail, err = image.SaveJpeg(jo)
	case vips.ImageTypePNG:
		po := vips.NewPngSaveOptions()
		po.Quantize = true
		if q.Quality == 100 {
			po.Quantize = false
		}
		thumbnail, err = image.SavePng(po)
		if q.Iterations > 0 {
			later = t.laterOptimizePng
		}
	case vips.ImageTypeWEBP:
		wo := vips.NewWebpSaveOptions()
		wo.Quality = int(q.Quality)
		thumbnail, err = image.SaveWebp(wo)
	case vips.ImageTypeAVIF:
		ao := vips.NewAvifSaveptions()
		ao.Quality = int(q.Quality)
		thumbnail, err = image.SaveAvif(ao)
	default:
		thumbnail, err = image.Save(options.ImageType(), int(q.Quality))
	}

	return
}

func (t *thumbnailer) transformFrames(imageId uint64, origin []byte, image *vips.Image, originType vips.ImageType, options *ThumbnailOptions) (err error) {
	if options.Trim() {
		options.SetTrim(false)
		t.logger.Warn("Trim is not supported for animated images")
	}

	imgWidth := image.Width()

	frameHeight, err := image.GetInt("page-height")
	if err != nil {
		return err
	}

	framesCount := image.Height() / frameHeight
	if nPages, _ := image.GetIntDefault("n-pages", 0); nPages > framesCount {
		if err = image.Load(origin, originType, 1, 1.0, framesCount); err != nil {
			return err
		}
	}

	t.logger.Info(fmt.Sprintf("[thumbnailer] #%d is animated with %d frames", imageId, framesCount))

	delay, err := image.GetIntSliceDefault("delay", nil)
	if err != nil {
		return err
	}
	loop, err := image.GetIntDefault("loop", 0)
	if err != nil {
		return err
	}
	gifLoop, err := image.GetIntDefault("gif-loop", -1)
	if err != nil {
		return err
	}
	gifDelay, err := image.GetIntDefault("gif-delay", -1)
	if err != nil {
		return err
	}

	frames := make([]*vips.Image, framesCount)
	defer func() {
		for _, frame := range frames {
			if frame != nil {
				frame.Clear()
			}
		}
	}()

	for i := 0; i < framesCount; i++ {
		frame := new(vips.Image)

		if err = image.Extract(frame, 0, i*frameHeight, imgWidth, frameHeight); err != nil {
			return err
		}

		frames[i] = frame

		if err = t.transformFrame(imageId, frame, options); err != nil {
			return err
		}

		if err = frame.CopyMemory(); err != nil {
			return err
		}
	}

	if err = image.Arrayjoin(frames); err != nil {
		return err
	}
	if err = image.CastUchar(); err != nil {
		return err
	}
	if err = image.CopyMemory(); err != nil {
		return err
	}

	if len(delay) == 0 {
		delay = make([]int, framesCount)
		for i := range delay {
			delay[i] = 40
		}
	} else if len(delay) > framesCount {
		delay = delay[:framesCount]
	}

	image.SetInt("page-height", frames[0].Height())
	image.SetIntSlice("delay", delay)
	image.SetInt("loop", loop)
	image.SetInt("n-pages", framesCount)

	if gifLoop >= 0 {
		image.SetInt("gif-loop", gifLoop)
	}
	if gifDelay >= 0 {
		image.SetInt("gif-delay", gifDelay)
	}

	return
}

func (t *thumbnailer) transformFrame(imageId uint64, image *vips.Image, options *ThumbnailOptions) (err error) {
	whiteColor := vips.RgbColor{
		R: 255,
		G: 255,
		B: 255,
	}
	trimmed := false
	flattened := false

	if err = image.Rad2Float(); err != nil {
		return err
	}
	if err = image.RgbColourspace(); err != nil {
		return err
	}

	// set transparent background
	if options.TransparentBackground() && !options.OpaqueBackground() && options.ImageType().SupportsAlpha() {
		t.logger.Info(fmt.Sprintf("[thumbnailer] #%d set transparent background", imageId))
		err = createTransparentBackground(image)
		if err != nil {
			return
		}
	}

	// set opaque background
	if options.OpaqueBackground() || image.HasAlpha() && !options.ImageType().SupportsAlpha() {
		t.logger.Info(fmt.Sprintf("[thumbnailer] #%d flatten to white", imageId))
		flattened = true
		err = image.Flatten(whiteColor)
		if err != nil {
			return
		}
	}

	// trim
	if options.Trim() {
		t.logger.Info(fmt.Sprintf("[thumbnailer] #%d smart trim", imageId))
		err = image.Trim(
			10,
			true,
			whiteColor,
			false,
			false,
		)
		if err != nil {
			return
		}
		trimmed = true
	}

	// resize
	forceExtent := false
	shouldResize := false

	resizeWidth := options.Width()
	resizeHeight := options.Height()

	if options.Width() == 0 && options.Height() == 0 {
		resizeWidth = uint(image.Width())
		resizeHeight = uint(image.Height())
	} else if options.Width() == 0 {
		shouldResize = true
		resizeWidth = uint(math.Floor(float64(image.Width()) * float64(options.Height()) / float64(image.Height())))
	} else if options.Height() == 0 {
		shouldResize = true
		resizeHeight = uint(math.Floor(float64(image.Height()) * float64(options.Width()) / float64(image.Width())))
	} else if options.ResizeTensile() {
		shouldResize = true
	} else if options.ResizePrecise() {
		shouldResize = true
		resizeWidth, resizeHeight, err = calculateWHWithAspectRatio(uint(image.Width()), uint(image.Height()), options.Width(), options.Height(), true)
		if err != nil {
			return
		}
	} else if options.ResizeInverse() {
		shouldResize = true
		resizeWidth, resizeHeight, err = calculateWHWithAspectRatio(uint(image.Width()), uint(image.Height()), options.Width(), options.Height(), false)
		if err != nil {
			return
		}
	} else {
		forceExtent = true
	}

	padding := config.Get().Padding
	if options.Padding() && options.Trim() && resizeWidth > 2*padding && resizeHeight > 2*padding {
		t.logger.Info(fmt.Sprintf("[thumbnailer] #%d add padding", imageId))
		resizeWidth -= 2 * padding
		resizeHeight -= 2 * padding
		forceExtent = true
	}

	if shouldResize && (uint(image.Width()) != resizeWidth || uint(image.Height()) != resizeHeight) || trimmed {
		t.logger.Info(fmt.Sprintf("[thumbnailer] #%d resize to %dx%d", imageId, resizeWidth, resizeHeight))
		err = image.Thumbnail(int(resizeWidth), int(resizeHeight))
		if err != nil {
			return
		}
	}

	if forceExtent || options.Extent() {
		t.logger.Info(fmt.Sprintf("[thumbnailer] #%d extent image to %dx%d with transparent background: %t", imageId, options.Width(), options.Height(), options.ImageType().SupportsAlpha()))

		offX := (int(options.Width()) - image.Width()) / 2
		offY := (int(options.Height()) - image.Height()) / 2

		err = image.Embed(
			int(options.Width()),
			int(options.Height()),
			offX,
			offY,
			whiteColor,
			!flattened && options.ImageType().SupportsAlpha())
		if err != nil {
			return
		}
	}

	if err = image.TransformColourProfile(); err != nil {
		return err
	}
	if err = image.RemoveColourProfile(); err != nil {
		return err
	}
	if err = image.CastUchar(); err != nil {
		return err
	}
	if err = image.Strip(); err != nil {
		return err
	}
	if err = image.CopyMemory(); err != nil {
		return err
	}

	return nil
}

func (t *thumbnailer) getQuality(width, height uint, imgtype vips.ImageType) quality {
	defaultQuality := quality{
		Quality:    config.Get().QualityDefault,
		Iterations: 100,
	}

	imgTypeExtension := imgtype.String()

	result := defaultQuality
	halfPerimeter := width + height

	for _, qualityFormat := range config.Get().Quality {
		if qualityFormat.Format == imgTypeExtension {
			result.Quality = qualityFormat.Quality
			result.Iterations = qualityFormat.Iterations

			for _, condition := range qualityFormat.Conditions {
				if halfPerimeter >= condition.From && halfPerimeter < condition.To {
					result.Quality = condition.Quality
					result.Iterations = condition.Iterations
					break
				}
			}
			break
		}
	}

	if result.Quality <= 0 {
		result = defaultQuality
	}
	if result.Quality > 100 {
		result.Quality = 100
	}
	if result.Iterations < 0 {
		result.Iterations = 100
	}

	return result
}

func (t *thumbnailer) laterOptimizePng(uncompressed []byte) (compressed []byte, err error) {
	originMime := fileinfo.MimeByData(uncompressed)
	originType := vips.ImageTypeByByMime(originMime)
	if originType == vips.ImageTypeUnknown {
		err = errors.New("unknown origin image type")
		return
	}

	image := new(vips.Image)
	defer image.Clear()
	defer vips.Cleanup()

	err = image.Load(uncompressed, originType, 1, 1.0, 1)
	if err != nil {
		return
	}

	quality := t.getQuality(uint(image.Width()), uint(image.Height()), originType)

	uncompressedFile, err := ioutil.TempFile("", "thumbnail-pngi")
	defer func(name string) {
		_ = os.Remove(name)
	}(uncompressedFile.Name())

	err = ioutil.WriteFile(uncompressedFile.Name(), uncompressed, 0644)
	if err != nil {
		return
	}

	compressedFile, err := ioutil.TempFile("", "thumbnail-pngo")
	defer func(name string) {
		_ = os.Remove(name)
	}(compressedFile.Name())

	_, err = helper.Exec("zopflipng", "--iterations="+strconv.FormatUint(uint64(quality.Iterations), 10), "-y", "--filters=01234mepb", "--lossy_8bit", "--lossy_transparent", uncompressedFile.Name(), compressedFile.Name())
	if err != nil {
		return
	}

	compressed, err = ioutil.ReadFile(compressedFile.Name())
	if err != nil {
		return
	}

	return
}

func (t *thumbnailer) newImageId() uint64 {
	t.imageId++
	return t.imageId
}

func NewThumbnailer(logger logging.Logger) Thumbnailer {
	result := &thumbnailer{}
	result.logger = logger
	return result
}
