package asset

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ItemStatusOnSale  = "on_sale"
	ItemStatusTrading = "trading"
	ItemStatusSoldOut = "sold_out"
	ItemStatusStop    = "stop"
	ItemStatusCancel  = "cancel"

	TransactionEvidenceStatusWaitShipping = "wait_shipping"
	TransactionEvidenceStatusWaitDone     = "wait_done"
	TransactionEvidenceStatusDone         = "done"

	ShippingsStatusInitial    = "initial"
	ShippingsStatusWaitPickup = "wait_pickup"
	ShippingsStatusShipping   = "shipping"
	ShippingsStatusDone       = "done"

	ItemsPerPage = 48

	ActiveSellerNumSellItems = 100
)

type AppUser struct {
	ID           int64  `json:"id"`
	AccountName  string `json:"account_name"`
	Password     string `json:"plain_passwd"`
	Address      string `json:"address,omitempty"`
	NumSellItems int    `json:"num_sell_items"`
}

type AppItem struct {
	ID          int64  `json:"id"`
	SellerID    int64  `json:"seller_id"`
	BuyerID     int64  `json:"buyer_id"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	ImageName   string `json:"image_name"`
	CategoryID  int    `json:"category_id"`
	CreatedAt   int64  `json:"created_at"`
}

type AppCategory struct {
	ID                 int    `json:"id"`
	ParentID           int    `json:"parent_id"`
	CategoryName       string `json:"category_name"`
	ParentCategoryName string `json:"parent_category_name,omitempty"`
}

type AppTransactionEvidence struct {
	ID                 int64  `json:"id"`
	SellerID           int64  `json:"seller_id"`
	BuyerID            int64  `json:"buyer_id"`
	Status             string `json:"status"`
	ItemID             int64  `json:"item_id"`
	ItemName           string `json:"item_name"`
	ItemPrice          int    `json:"item_price"`
	ItemDescription    string `json:"item_description"`
	ItemCategoryID     int    `json:"item_category_id"`
	ItemRootCategoryID int    `json:"item_root_category_id"`
	CreatedAt          int64  `json:"created_at"`
	UpdatedAt          int64  `json:"updated_at"`
}

type ImageMD5 struct {
	Name string `json:"name"`
	MD5  string `json:"md5"`
}

var (
	users                map[int64]AppUser
	activeSellerIDs      []int64
	buyerIDs             []int64
	items                map[string]AppItem
	categories           map[int]AppCategory
	rootCategories       []AppCategory
	childCategories      []AppCategory
	userItems            map[int64][]int64
	transactionEvidences map[int64]AppTransactionEvidence
	keywords             []string
	imageFiles           []string
	imageMD5Lists        map[string]string
	muItem               sync.RWMutex
	muUser               sync.RWMutex
	muImageFile          sync.Mutex
	indexImageFile       int
	indexActiveSellerID  int32
	indexBuyerID         int32
)

// Initialize is a function to load initial data
func Initialize(dataDir string) {
	users = make(map[int64]AppUser)
	activeSellerIDs = make([]int64, 0, 400)
	buyerIDs = make([]int64, 0, 1000)
	items = make(map[string]AppItem)
	categories = make(map[int]AppCategory)
	rootCategories = make([]AppCategory, 0, 10)
	childCategories = make([]AppCategory, 0, 50)
	userItems = make(map[int64][]int64)
	transactionEvidences = make(map[int64]AppTransactionEvidence)
	imageFiles = make([]string, 0, 10000)
	imageMD5Lists = make(map[string]string)

	f, err := os.Open(filepath.Join(dataDir, "result/users_json.txt"))
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	user := &AppUser{}

	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), user)
		if err != nil {
			log.Fatal(err)
		}
		users[user.ID] = *user

		if user.NumSellItems >= ActiveSellerNumSellItems {
			activeSellerIDs = append(activeSellerIDs, user.ID)
		} else {
			buyerIDs = append(buyerIDs, user.ID)
		}
	}
	f.Close()

	f, err = os.Open(filepath.Join(dataDir, "result/items_json.txt"))
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(f)
	item := AppItem{}

	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &item)
		if err != nil {
			log.Fatal(err)
		}
		items[fmt.Sprintf("%d_%d", item.SellerID, item.ID)] = item
		if item.Status == ItemStatusOnSale {
			userItems[item.SellerID] = append(userItems[item.SellerID], item.ID)
		}
	}
	f.Close()

	f, err = os.Open(filepath.Join(dataDir, "result/category_json.txt"))
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(f)
	category := AppCategory{}

	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &category)
		if err != nil {
			log.Fatal(err)
		}
		categories[category.ID] = category

		if category.ParentID == 0 {
			rootCategories = append(rootCategories, category)
		} else {
			childCategories = append(childCategories, category)
		}
	}
	f.Close()

	f, err = os.Open(filepath.Join(dataDir, "result/transaction_evidences_json.txt"))
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(f)
	te := AppTransactionEvidence{}

	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &te)
		if err != nil {
			log.Fatal(err)
		}
		transactionEvidences[te.ID] = te
	}
	f.Close()

	f, err = os.Open(filepath.Join(dataDir, "image_files_md5_json.txt"))
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(f)
	im := ImageMD5{}

	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &im)
		if err != nil {
			log.Fatal(err)
		}
		imageMD5Lists[im.Name] = im.MD5
	}
	f.Close()

	f, err = os.Open(filepath.Join(dataDir, "keywords.tsv"))
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(f)

	for scanner.Scan() {
		text := scanner.Text()
		keywords = append(keywords, text)
	}
	f.Close()

	d, err := os.Open(filepath.Join(dataDir, "images"))
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		imageFiles = append(imageFiles, filepath.Join(dataDir, "images", file.Name()))
	}

	rand.Shuffle(len(activeSellerIDs), func(i, j int) { activeSellerIDs[i], activeSellerIDs[j] = activeSellerIDs[j], activeSellerIDs[i] })
	rand.Shuffle(len(buyerIDs), func(i, j int) { buyerIDs[i], buyerIDs[j] = buyerIDs[j], buyerIDs[i] })
	rand.Shuffle(len(imageFiles), func(i, j int) { imageFiles[i], imageFiles[j] = imageFiles[j], imageFiles[i] })
}

func (u1 *AppUser) Equal(u2 *AppUser) bool {
	return u1.AccountName == u2.AccountName && u1.Address == u2.Address
}

func GetRandomActiveSeller() AppUser {
	muUser.RLock()
	defer muUser.RUnlock()
	// 全部使い切ったらpanicするので十分なユーザー数を用意しておく
	return users[activeSellerIDs[len(activeSellerIDs)-int(atomic.AddInt32(&indexActiveSellerID, 1))]]
}

func GetRandomActiveSellerIDs(num int) []int64 {
	len := len(activeSellerIDs)
	if num > len {
		num = len
	}
	newIDs := make([]int64, 0, num)
	s := rand.Intn(len)
	for i := 0; i < num; i++ {
		newIDs = append(newIDs, activeSellerIDs[s])
		s++
		if s == len {
			s = 0
		}
	}
	return newIDs
}

func GetRandomBuyer() AppUser {
	muUser.RLock()
	defer muUser.RUnlock()
	// 全部使い切ったらpanicするので十分なユーザー数を用意しておく
	return users[buyerIDs[len(buyerIDs)-int(atomic.AddInt32(&indexBuyerID, 1))]]
}

func GetRandomBuyerIDs(num int) []int64 {
	len := len(buyerIDs)
	if num > len {
		num = len
	}
	newIDs := make([]int64, 0, num)
	s := rand.Intn(len)
	for i := 0; i < num; i++ {
		newIDs = append(newIDs, buyerIDs[s])
		s++
		if s == len {
			s = 0
		}
	}
	return newIDs
}

func GetUser(sellerID int64) AppUser {
	muUser.RLock()
	defer muUser.RUnlock()
	return users[sellerID]
}

func GetUserItemsFirst(sellerID int64) int64 {
	muItem.RLock()
	defer muItem.RUnlock()

	return userItems[sellerID][0]
}

func GetUserItems(sellerID int64) []int64 {
	muItem.RLock()
	defer muItem.RUnlock()

	return userItems[sellerID]
}

func GetImageMD5(imageURL string) (md5Str string) {
	return imageMD5Lists[imageURL]
}

func GetItem(sellerID, itemID int64) (AppItem, bool) {
	i, ok := getItem(sellerID, itemID)
	for j := 1; !ok && j < 1025; j = j * 2 {
		<-time.After(time.Duration(j) * time.Millisecond)
		i, ok = getItem(sellerID, itemID)
	}
	return i, ok
}

func getItem(sellerID, itemID int64) (AppItem, bool) {
	muItem.RLock()
	defer muItem.RUnlock()

	i, ok := items[fmt.Sprintf("%d_%d", sellerID, itemID)]
	return i, ok
}

func SetItem(sellerID int64, itemID int64, name string, price int, description string, categoryID int) {
	muItem.Lock()
	defer muItem.Unlock()
	muUser.Lock()
	defer muUser.Unlock()

	userItems[sellerID] = append(userItems[sellerID], itemID)

	key := fmt.Sprintf("%d_%d", sellerID, itemID)
	items[key] = AppItem{
		ID:          itemID,
		SellerID:    sellerID,
		Status:      ItemStatusOnSale,
		Name:        name,
		Price:       price,
		Description: description,
		CategoryID:  categoryID,
		CreatedAt:   time.Now().Unix(),
	}

	user := users[sellerID]
	user.NumSellItems = user.NumSellItems + 1
	users[sellerID] = user
}

func SetItemPrice(sellerID int64, itemID int64, price int) {
	muItem.Lock()
	defer muItem.Unlock()

	key := fmt.Sprintf("%d_%d", sellerID, itemID)
	item := items[key]
	item.Price = price

	items[key] = item
}

func SetItemCreatedAt(sellerID int64, itemID int64, createdAt int64) {
	muItem.Lock()
	defer muItem.Unlock()

	key := fmt.Sprintf("%d_%d", sellerID, itemID)
	item := items[key]
	item.CreatedAt = createdAt

	items[key] = item
}

func GetRandomImageFileName() string {
	muImageFile.Lock()
	defer muImageFile.Unlock()

	indexImageFile--

	if indexImageFile < 0 {
		indexImageFile = len(imageFiles) - 1
	}

	return imageFiles[indexImageFile]
}

func GetRandomRootCategory() AppCategory {
	return rootCategories[rand.Intn(len(rootCategories))]
}

func GetRootCategories() []AppCategory {
	return rootCategories
}

func GetRandomChildCategory() AppCategory {
	return childCategories[rand.Intn(len(childCategories))]
}

func GetCategory(categoryID int) (AppCategory, bool) {
	c, ok := categories[categoryID]
	return c, ok
}

// TODO: transactionEvidencesをちゃんと管理するようにして存在しないケースをなくす
func GetTransactionEvidence(id int64) (AppTransactionEvidence, bool) {
	te, ok := transactionEvidences[id]
	return te, ok
}

func GenText(length int, isLine bool) string {
	texts := make([]string, 0, length)

	for i := 0; i < length; i++ {
		t := keywords[rand.Intn(len(keywords))]

		if t == "#" {
			if isLine {
				t = "\n"
			} else {
				t = " "
			}
		}

		texts = append(texts, t)
	}

	return strings.Join(texts, "")
}
