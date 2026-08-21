package wawi

import (
	"net/http"
	"testing"
)

const staffelErr = `{"ErrorCode":"ValidationError","ValidationErrors":{},"Errors":{},"ErrorMessage":"- StaffelPreisNotFound","Stacktrace":"   bei JTL.Wawi.ArtikelVerwaltung.Core.Artikeldetails.Preise.ArtikelPreiseAggregate.UpdateShopPreis(ShopKey shopKey)\r\n   bei System.Web.Http.HttpServer.<SendAsync>d__24.MoveNext()"}`

func TestIsMissingPriceRow(t *testing.T) {
	if !isMissingPriceRow(http.StatusBadRequest, []byte(staffelErr)) {
		t.Error("StaffelPreisNotFound was not recognised as a missing price row")
	}
	if isMissingPriceRow(http.StatusBadRequest, []byte(`{"ErrorMessage":"- InvalidCustomerGroup"}`)) {
		t.Error("an unrelated validation error must not trigger a create")
	}
	if isMissingPriceRow(http.StatusInternalServerError, []byte(staffelErr)) {
		t.Error("only 400 and 404 may trigger a create")
	}
}

func TestAPIErrorTextDropsStacktrace(t *testing.T) {
	got := apiErrorText([]byte(staffelErr))
	want := "ValidationError: StaffelPreisNotFound"
	if got != want {
		t.Errorf("apiErrorText() = %q, want %q", got, want)
	}
}
