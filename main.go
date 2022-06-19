package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type Product struct {
	Name        string  `json:"name"`
	Image       string  `json:"image_url"`
	Price       string  `json:"price"`
	Rating      float64 `json:"rating"`
	Merchant    string
	Description string
}

func parseResult(result []interface{}) []*Product {
	products := []*Product{}
	for _, val := range result {
		data := val.(map[string]interface{})["data"].(map[string]interface{})["CategoryProducts"].(map[string]interface{})["data"]

		for _, value := range data.([]interface{}) {
			product := &Product{}
			product.Name = value.(map[string]interface{})["name"].(string)
			product.Image = value.(map[string]interface{})["imageUrl"].(string)
			product.Price = value.(map[string]interface{})["price"].(string)
			product.Rating = value.(map[string]interface{})["rating"].(float64)
			product.Merchant = value.(map[string]interface{})["shop"].(map[string]interface{})["name"].(string)
			products = append(products, product)
		}
	}
	return products
}

func scrape(page int) *http.Response {
	pageStr := strconv.Itoa(page)
	var payload = []byte(`[{"operationName":"SearchProductQuery","variables":{"params":"page=` + pageStr + `&ob=&identifier=handphone-tablet_handphone&sc=24&user_id=0&rows=60&start=61&source=directory&device=desktop&page=2&related=true&st=product&safe_search=false","adParams":"page=2&page=2&dep_id=24&ob=&ep=product&item=15&src=directory&device=desktop&user_id=0&minimum_item=15&start=61&no_autofill_range=5-14"},"query":"query SearchProductQuery($params: String, $adParams: String) {\n  CategoryProducts: searchProduct(params: $params) {\n    count\n    data: products {\n      id\n      url\n      imageUrl: image_url\n      imageUrlLarge: image_url_700\n      catId: category_id\n      gaKey: ga_key\n      countReview: count_review\n      discountPercentage: discount_percentage\n      preorder: is_preorder\n      name\n      price\n      original_price\n      rating\n      wishlist\n      labels {\n        title\n        color\n        __typename\n      }\n      badges {\n        imageUrl: image_url\n        show\n        __typename\n      }\n      shop {\n        id\n        url\n        name\n        goldmerchant: is_power_badge\n        official: is_official\n        reputation\n        clover\n        location\n        __typename\n      }\n      labelGroups: label_groups {\n        position\n        title\n        type\n        __typename\n      }\n      __typename\n    }\n    __typename\n  }\n  displayAdsV3(displayParams: $adParams) {\n    data {\n      id\n      ad_ref_key\n      redirect\n      sticker_id\n      sticker_image\n      productWishListUrl: product_wishlist_url\n      clickTrackUrl: product_click_url\n      shop_click_url\n      product {\n        id\n        name\n        wishlist\n        image {\n          imageUrl: s_ecs\n          trackerImageUrl: s_url\n          __typename\n        }\n        url: uri\n        relative_uri\n        price: price_format\n        campaign {\n          original_price\n          discountPercentage: discount_percentage\n          __typename\n        }\n        wholeSalePrice: wholesale_price {\n          quantityMin: quantity_min_format\n          quantityMax: quantity_max_format\n          price: price_format\n          __typename\n        }\n        count_talk_format\n        countReview: count_review_format\n        category {\n          id\n          __typename\n        }\n        preorder: product_preorder\n        product_wholesale\n        free_return\n        isNewProduct: product_new_label\n        cashback: product_cashback_rate\n        rating: product_rating\n        top_label\n        bottomLabel: bottom_label\n        __typename\n      }\n      shop {\n        image_product {\n          image_url\n          __typename\n        }\n        id\n        name\n        domain\n        location\n        city\n        tagline\n        goldmerchant: gold_shop\n        gold_shop_badge\n        official: shop_is_official\n        lucky_shop\n        uri\n        owner_id\n        is_owner\n        badges {\n          title\n          image_url\n          show\n          __typename\n        }\n        __typename\n      }\n      applinks\n      __typename\n    }\n    template {\n      isAd: is_ad\n      __typename\n    }\n    __typename\n  }\n}\n"}]`)

	req, err := http.NewRequest("POST", "https://gql.tokopedia.com/graphql/SearchProductQuery", bytes.NewBuffer(payload))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("origin", "https://www.tokopedia.com")

	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func main() {
	page := 1
	products := []*Product{}

	for i := 1; len(products) <= 100; i++ {
		// if page >= 2 {
		// 	break
		// }
		resp := scrape(page)
		body, _ := ioutil.ReadAll(resp.Body)
		data := []interface{}{}

		err := json.Unmarshal(body, &data)
		if err != nil {
			panic(err)
		}
		result := parseResult(data)

		page += 1
		products = append(products, result...)
	}

	csvFile, err := os.Create("data.csv")
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	cw := csv.NewWriter(csvFile)

	cw.Write([]string{"no", "name", "image_url", "price", "rating", "merchant"})
	for i, product := range products {
		var row []string
		row = append(row, strconv.Itoa(i+1))
		row = append(row, product.Name)
		row = append(row, product.Image)
		row = append(row, product.Price)
		row = append(row, strconv.FormatFloat(product.Rating, 'f', 2, 64))
		row = append(row, product.Merchant)
		cw.Write(row)
		if i == 99 {
			break
		}
	}
	defer cw.Flush()
}
