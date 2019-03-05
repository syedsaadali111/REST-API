package main
import (
  "net/http"
  "strconv"
  "encoding/json"
  "os"
  "fmt"
  "io/ioutil"
  "log"
  "math/rand"
  "time"
)

type Games struct {
  Games []Game `json:"games"`
}

type Game struct {
  Session_id int `json:"id"`
  Max_round int `json:"max_round"`
  Last_round int `json:"last_round"`
  WinCount int `json:"win"`
  LoseCount int `json:"lose"`
  Past_array []Past_data `json:"past_data"`
}

type Past_data struct {
  Choose string `json:"choose"`
  Me string `json:"me"`
}

func readFile(allGames *Games, jsonFile *os.File) {

  byteValue, _ := ioutil.ReadAll(jsonFile)

  json.Unmarshal(byteValue, allGames)
}

func writeFile(allGames Games, jsonFile *os.File){

  jsonFile.Close()              //close existing file

  err := os.Remove("games.json")  //remove existing file
  if err != nil {
      log.Fatal(err)
  }

  newFile, err := os.Create("games.json") //create a new file
  if err != nil {
      log.Fatal(err)
  }
  log.Println(newFile)
  newFile.Close()

  newJsonFile, err := os.OpenFile("games.json", os.O_RDWR, 0644)

  newBytes, err := json.MarshalIndent(allGames, "", "    ")  //prepare new data
	if err != nil {
		panic(err)
  }

  _, err = newJsonFile.WriteAt(newBytes, 0)  //write at new file
	if err != nil {
		panic(err)
  }
  newJsonFile.Close()
  
}

func newGame(w http.ResponseWriter, r *http.Request) {


  //GET FILE DATA

  var allGames Games

  jsonFile, err := os.OpenFile("games.json", os.O_RDWR, 0644)

  if err != nil {
    fmt.Println(err)
  }

  fmt.Println("File opened successful")

  defer jsonFile.Close()

  readFile(&allGames, jsonFile)

  //GENERATE ID MESSAGE
  var id int

  if (len(allGames.Games) == 0) {
    id = 1 
  } else {
    id = allGames.Games[len(allGames.Games) - 1].Session_id + 1
  }

  message := "New Rock-Paper-Scissors Game Started\nSession ID: " + strconv.Itoa(id)

  //CREATE NEW GAME
  var newGame Game

  keys, ok := r.URL.Query()["round"]
    
  if !ok || len(keys[0]) < 1 {          //if round parameter is not given, set to 1 as default
      log.Println("Url Param 'round' is missing")
      message += "\n\nNumber of rounds not specified, default value is 1"
      newGame.Max_round = 1
  } else {
      key, _ := strconv.Atoi(keys[0])
      newGame.Max_round = key
  }

  message += "\n\n\nto play Rock: http://localhost:9999/play?id=" + strconv.Itoa(id) + "&choose=rock\nto play Paper: http://localhost:9999/play?id=" + strconv.Itoa(id) + "&choose=paper\nto play Scissors: http://localhost:9999/play?id=" + strconv.Itoa(id) + "&choose=scissors"

  w.Write([]byte (message))

  newGame.Session_id = id           //initializing new game object
  newGame.Last_round = 0
  newGame.Past_array = []Past_data{}
  newGame.WinCount = 0
  newGame.LoseCount = 0

  //APPEND AND OVERWRITE

  allGames.Games = append(allGames.Games, newGame)

  writeFile(allGames, jsonFile)


}

func play(w http.ResponseWriter, r *http.Request) {
  message:= ""
  
  keys, ok := r.URL.Query()["id"]
    
    if !ok || len(keys[0]) < 1 {                  //check if id parameter is present
        message += "Session id is missing"
        w.Write([]byte (message))
        return
    }

    id, _ := strconv.Atoi(keys[0]) //id param store

    keys, ok = r.URL.Query()["choose"]
    
    if !ok || len(keys[0]) < 1 {                  //check if choose parameter is present
        message += "Choice parameter is missing"
        w.Write([]byte (message))
        return
    }

    choose := keys[0] //choose param store

    //READ FILE

    var allGames Games
    
    jsonFile, err := os.OpenFile("games.json", os.O_RDWR, 0644)

    if err != nil {
      fmt.Println(err)
    }

    fmt.Println("File opened successful")

    readFile(&allGames, jsonFile)

    jsonFile.Close()

    //VALIDATE SESSION ID

    i := 0
    var index int
    found := false
    for i < (len(allGames.Games)) || found == true {
      if allGames.Games[i].Session_id == id {
        found = true
        index = i
        break
      }
      i++
    }

    if (found == false) {
      message += "Session id is invalid"
    } else {
      message += "-> ROUND " + strconv.Itoa(allGames.Games[index].Last_round + 1) + " of " + strconv.Itoa(allGames.Games[index].Max_round)
      
      allGames.Games[index].Last_round += 1       //inc last round value
      
      choiceMap := map[string]int{"rock": 1, "paper": 2, "scissors": 3}
      chooseInt, check := choiceMap[choose]     //get choice as int
      
      if (check == false) {
        message += "Invalid input for choose"
      } else {
        s1 := rand.NewSource(time.Now().UnixNano())  //generating random choice
        r1 := (rand.New(s1))
        myChoice := r1.Intn(3) + 1
        
        var record Past_data

        if (myChoice == 1) {              //display and store game data
          message += "\n\nMe: " + "rock"
          record.Me = "rock"
        } else if (myChoice == 2) {
          message += "\n\nMe: " + "paper"
          record.Me = "paper"
        } else {
          message += "\n\nMe: " + "scissors"
          record.Me = "scissors"
        }

        message += "\nYou: " + choose
        record.Choose = choose

        if (chooseInt == myChoice) {                //evaluatiing the winner
          message += "\n\nITS A DRAW"
        } else if (chooseInt - (myChoice % 3) == 1) {
          message += "\n\nYOU WIN"
          allGames.Games[index].WinCount++
        } else {
          message += "\n\nYOU LOST"
          allGames.Games[index].LoseCount++
        }

        allGames.Games[index].Past_array = append(allGames.Games[index].Past_array, record)  //updating session history

        if (allGames.Games[index].Last_round == allGames.Games[index].Max_round) {  //if this was the last round..
          message += "\n\n-> GAME COMPLETED"

          for i := 0; i < len(allGames.Games[index].Past_array); i++ {
            message += "\nRound " + strconv.Itoa(i+1) + ": " + allGames.Games[index].Past_array[i].Me + " vs. " +  allGames.Games[index].Past_array[i].Choose
          }

          message += "\nMe vs. You\n" + strconv.Itoa(allGames.Games[index].LoseCount) + " vs. " + strconv.Itoa(allGames.Games[index].WinCount)

          if (strconv.Itoa(allGames.Games[index].LoseCount) == strconv.Itoa(allGames.Games[index].WinCount)) {
            message += "\n\nGAME IS A DRAW!!!"
          } else if (strconv.Itoa(allGames.Games[index].LoseCount) < strconv.Itoa(allGames.Games[index].WinCount)) {
            message += "\n\nYOU WON!!!"
          } else {
            message += "\n\nYOU LOST!!!"
          }

          allGames.Games[index] = allGames.Games[ len(allGames.Games) - 1]  //remove session from json file
          allGames.Games = allGames.Games[:len(allGames.Games)-1]
          fmt.Println(allGames.Games)

        } else {
          message += "\n\n" + strconv.Itoa(allGames.Games[index].Max_round - allGames.Games[index].Last_round) + " round(s) left"
        }
        writeFile(allGames, jsonFile) //update the file
      }

    }

    w.Write([]byte (message))  //display output
}

func main() {
  http.HandleFunc("/newGame", newGame)
  http.HandleFunc("/play", play)
  if err := http.ListenAndServe(":9999", nil); err != nil {
    panic(err)
  }
}