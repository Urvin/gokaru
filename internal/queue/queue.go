package queue

import (
	"github.com/urvin/gokaru/internal/contracts"
	strg "github.com/urvin/gokaru/internal/storage"
	thmbnlr "github.com/urvin/gokaru/internal/thumbnailer"
	"log/slog"
	"sync"
	"time"
)

type entry struct {
	ready     chan struct{}
	err       error
	thumbnail contracts.FileDto
	miniature *contracts.MiniatureDto
}

type later struct {
	fn        func([]byte) ([]byte, error)
	miniature contracts.MiniatureDto
}

type Queue struct {
	entriesMx sync.Mutex
	entries   map[string]*entry

	entriesProcs chan *entry
	latersProcs  chan later

	logger      *slog.Logger
	storage     strg.Storage
	thumbnailer thmbnlr.Thumbnailer
}

func NewQueue(logger *slog.Logger, storage strg.Storage, thumbnailer thmbnlr.Thumbnailer, procs uint, postProcs uint) *Queue {
	q := &Queue{
		entries:      make(map[string]*entry),
		logger:       logger,
		storage:      storage,
		thumbnailer:  thumbnailer,
		entriesProcs: make(chan *entry, procs),
		latersProcs:  make(chan later, postProcs),
	}
	var i uint
	for i = 0; i < procs; i++ {
		go q.processEntries()
	}
	for i = 0; i < postProcs; i++ {
		go q.processLaters()
	}
	return q
}
func (q *Queue) GetThumbnail(miniature *contracts.MiniatureDto) (thumbnail contracts.FileDto, err error) {

	key := miniature.Hash()

	q.entriesMx.Lock()
	e := q.entries[key]

	if e == nil {
		e = &entry{
			ready:     make(chan struct{}),
			miniature: miniature,
		}
		q.entries[key] = e
		q.entriesMx.Unlock()

		q.entriesProcs <- e
		<-e.ready
	} else {
		q.entriesMx.Unlock()
		<-e.ready
	}

	err = e.err
	thumbnail = e.thumbnail

	q.entriesMx.Lock()
	delete(q.entries, key)
	q.entriesMx.Unlock()

	return
}

func (q *Queue) processEntries() {
	for e := range q.entriesProcs {
		e.thumbnail, e.err = q.obtainThumbnail(e.miniature)
		close(e.ready)
	}
}

func (q *Queue) obtainThumbnail(miniature *contracts.MiniatureDto) (thumbnail contracts.FileDto, err error) {
	if q.storage.ThumbnailExists(miniature) {
		thumbnail, err = q.storage.ReadThumbnail(miniature)
		return
	}
	thumbnail, err = q.processThumbnail(miniature)
	return
}

func (q *Queue) processThumbnail(miniature *contracts.MiniatureDto) (thumbnail contracts.FileDto, err error) {
	origin := contracts.OriginDto{
		Type:     miniature.Type,
		Category: miniature.Category,
		Name:     miniature.Name,
	}
	originInfo, err := q.storage.Read(&origin)
	if err != nil {
		return
	}

	options := thmbnlr.ThumbnailOptions{}
	options.SetWidth(uint(miniature.Width))
	options.SetHeight(uint(miniature.Height))
	options.SetImageTypeWithExtension(miniature.Extension)
	options.SetOptionsWithCast(uint(miniature.Cast))

	bytes, ltr, err := q.thumbnailer.Thumbnail(originInfo.Contents, options)

	if err != nil {
		return
	}

	err = q.storage.WriteThumbnail(miniature, bytes)
	if err != nil {
		return
	}

	thumbnail.Contents = bytes

	if ltr != nil {
		go func(ltr later) {
			q.latersProcs <- ltr
		}(later{
			fn:        ltr,
			miniature: *miniature,
		})
	}

	return
}

func (q *Queue) processLaters() {
	for ltr := range q.latersProcs {
		err := q.processLater(ltr)
		if err != nil {
			q.logger.Error(
				"Could not process later",
				"context", "queue",
				"handler", "processLaters",
				"error", err.Error(),
			)
		}
	}
}

func (q *Queue) processLater(ltr later) (err error) {
	start := time.Now()

	file, err := q.storage.ReadThumbnail(&ltr.miniature)
	if err != nil {
		return
	}
	data, err := ltr.fn(file.Contents)
	if err != nil {
		return
	}

	err = q.storage.WriteThumbnail(&ltr.miniature, data)

	q.logger.Info(
		"Postprocessed "+ltr.miniature.Hash()+" in "+time.Since(start).String(),
		"context", "queue",
		"handler", "processLater",
	)

	return
}
