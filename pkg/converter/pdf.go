package converter

import (
	"context"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ConvertHTMLToPDF преобразует HTML в PDF с использованием chromedp и cdproto/page.
// scale - масштаб (1.0 = 100%, 0.8 = 80%, 1.2 = 120%, и т.д.)
func ConvertHTMLToPDF(htmlContent string, scale float64) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	dataURL := "data:text/html;charset=utf-8," + url.PathEscape(htmlContent)
	var pdfBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(dataURL),
		chromedp.Sleep(2*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithDisplayHeaderFooter(true).
				WithMarginTop(0.4).
				WithMarginBottom(0.6). // Увеличиваем нижний отступ
				WithFooterTemplate(`
												  <div style="font-size:10px !important;
															  border-top:1px solid #888;
															  padding-top:5px;
															  width:90%;
															  margin:0 auto;
															  text-align:right;
															  color:#999;">
													Page <span class="pageNumber"></span> of <span class="totalPages"></span>
												  </div>
												`).
				WithHeaderTemplate(`<div></div>`).
				WithScale(scale).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, err
	}
	return pdfBuf, nil
}
