package middleout
// // assume there exists another file that handles the user requests
// // that file gets the ticker and range of dates and will call the functions in here to interact with the database

// // package just always seems to be declared as the director
// package middleout	// could I declare something to be a package from a different director?

// import (
// 	// "encoding/json"
// 	"fmt"
// 	"github.com/bluedresscapital/coattails/pkg/stockings"
// 	// "net/http"
// 	"time"
// 	"github.com/bluedresscapital/coattails/pkg/wardrobe"
// )




// func rangeInDB(ticker string, startDate time.Time, endDate time.Time) bool {

// }


// // lets assume we already have some table with every legal date
// // then when we get a start and end date, we want to find every legal date in between (inclusive), then we use those dates to check our database

// func findMarketDates(startDate time.Time, endDate time.Time) []int {
// 	// might be pointless to create a new function here, instead just make the function in DB file
// 	// for now I will just do the db stuff in this file



// }