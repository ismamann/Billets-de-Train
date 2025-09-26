package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Codes de couleur pour l'affichage dans le terminal
var Black string = "\033[0;30m"
var Red string = "\033[0;31m"
var Green string = "\033[0;32m"
var Yellow string = "\033[0;33m"
var Blue string = "\033[0;34m"
var Purple string = "\033[0;35m"
var White string = "\033[0;37m"

var BrightBlack string = "\033[1;30m"
var BrightRed string = "\033[1;31m"
var BrightGreen string = "\033[1;32m"
var BrightYellow string = "\033[1;33m"
var BrightBlue string = "\033[1;34m"
var BrightPurple string = "\033[1;35m"
var BrightCyan string = "\033[1;36m"
var BrightWhite string = "\033[1;37m"

var BgBlack string = "\033[40m"
var BgRed string = "\033[41m"
var BgGreen string = "\033[42m"
var BgYellow string = "\033[43m"
var BgBlue string = "\033[44m"
var BgPurple string = "\033[45m"
var BgCyan string = "\033[46m"
var BgWhite string = "\033[47m"

var Underline string = "\033[4m"
var Bold string = "\033[1m"
var Italic string = "\033[3m"

// Structure représentant un billet de train
// Contient les infos de base du train et le nombre de places disponibles
type Billet struct {
	numero_train int
	TrainID      string // Identifiant du train
	Depart       string
	Destination  string
	Date         time.Time
	//classe     int    // la classe de place : enum{1,2}
	PlacesRest  int // Nombre de places restantes
	CapaciteMax int // Capacité maximale du train
}

// Structure représentant une tentative d'achat de billet
type Achat_billet struct {
	Num_train int
	Nb        int
}

// Dictionnaire global pour les achats locaux
var Achat = make(map[int]Achat_billet)

func InitAchat() map[int]Achat_billet {
	Achat[1] = Achat_billet{Num_train: 1, Nb: 0}
	Achat[2] = Achat_billet{Num_train: 2, Nb: 0}
	Achat[3] = Achat_billet{Num_train: 3, Nb: 0}
	return Achat
}

// Stock global des billets (vue locale de l'application)
var Stock = make(map[int]Billet)

var (
	mu        sync.Mutex
	fileMutex sync.Mutex
)

// Initialisation du stock de billets (remplissage initial)
func InitStock() map[int]Billet {
	Stock[1] = Billet{
		numero_train: 1,
		TrainID:      "TGV2213",
		Depart:       "Paris",
		Destination:  "Lyon",
		Date:         time.Date(2025, time.July, 7, 0, 0, 0, 0, time.UTC),
		PlacesRest:   5,
		CapaciteMax:  20,
	}
	Stock[2] = Billet{
		numero_train: 2,
		TrainID:      "TGV1234",
		Depart:       "Marseille",
		Destination:  "Paris",
		Date:         time.Date(2024, time.June, 3, 0, 0, 0, 0, time.UTC),
		PlacesRest:   3,
		CapaciteMax:  20,
	}
	Stock[3] = Billet{
		numero_train: 3,
		TrainID:      "TGV6611",
		Depart:       "Lyon",
		Destination:  "Marseille",
		Date:         time.Date(2024, time.May, 21, 0, 0, 0, 0, time.UTC),
		PlacesRest:   7,
		CapaciteMax:  30,
	}

	return Stock
}

func AfficherTrain(trains map[int]Billet) {
	fmt.Println()
	fmt.Println(Underline + Bold + "Trains disponibles :" + Raz)
	for numero, train := range trains {
		fmt.Printf("%sTrain %s[%d]%s%s [%s] - %s - places restantes: %s%d%s%s - [%s --> %s]\n%s", Red, Bold, numero, Raz, Red, train.TrainID, train.Date.Format("02/01/2006"), Bold, train.PlacesRest, Raz, Red, train.Depart, train.Destination, Raz)
	}
}

// Retourner une chaîne décrivant l'état local des trains pour affichage ailleurs
func AfficherTrainEtatLocal(trains map[int]Billet, p_nom *string) string {
	str := ""
	str += ("Site " + *p_nom + " Trains disponibles :")
	for numero, train := range trains {
		str += fmt.Sprintf("Train [%d] [%s] - %s - places restantes: %d ++ ", numero, train.TrainID, train.Date.Format("02-01-2006"), train.PlacesRest)
	}
	return str
}

func AfficherMenu() {
	fmt.Println()
	fmt.Println(Underline + Bold + "Quelle action voulez-vous effectuer?" + Raz + Raz)
	fmt.Println("   " + Bold + "[1]" + Raz + " - Acheter un billet.")
	fmt.Println("   " + Bold + "[2]" + Raz + " - Annuler un billet.")
	fmt.Println("   " + Bold + "[3]" + Raz + " - Faire une sauvegarde des données.")
	fmt.Println("   " + Bold + "[0]" + Raz + " - Quitter.")
	fmt.Println()
}

func AfficherArret() {
	fmt.Println()
	fmt.Println()
	fmt.Println("############################")
	fmt.Println("############################")
	fmt.Println(Underline + Bold + "Merci de votre visite." + Raz + Raz)
	fmt.Println(Underline + Bold + "A bientot." + Raz + Raz)
	fmt.Println(Underline + Bold + "Arret de l'application dans:" + Raz + Raz)
	fmt.Println(rouge + "5s" + Raz)
	time.Sleep(1 * time.Second)
	fmt.Println(rouge + "4s" + Raz)
	time.Sleep(1 * time.Second)
	fmt.Println(rouge + "3s" + Raz)
	time.Sleep(1 * time.Second)
	fmt.Println(rouge + "2s" + Raz)
	time.Sleep(1 * time.Second)
	fmt.Println(rouge + "1s" + Raz)
	time.Sleep(1 * time.Second)
}

// Lire le choix de l'utilisateur dans le terminal
func ChoixUser() (int, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	for {
		fmt.Print(Underline + Bold + "Quel est votre choix :\n" + Raz + Raz)
		if scanner.Scan() {
			input := scanner.Text()
			choix, err := strconv.Atoi(input)
			if err != nil || choix < 0 || choix > 3 {
				fmt.Println("Choix invalide. Veuillez entrer un choix entre 0 et 3.")
				continue
			}
			return choix, nil
		}
		if err := scanner.Err(); err != nil {
			return 0, err
		}
	}
}

func GetChoixUser(msg string) (int, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	for {
		fmt.Print(msg)
		if scanner.Scan() {
			input := scanner.Text()
			numero, err := strconv.Atoi(input)
			if err != nil {
				fmt.Printf("Veuillez entrer un nombre entier.\n\n")
				continue
			}
			return numero, nil
		}
		if err := scanner.Err(); err != nil {
			return 0, err
		}
	}
}

// Vérifier si l'utilisateur veut annuler
func AnnulerChoix(choix int) bool {
	return choix == 0
}

// Les actions pour vérifier des informations sont valide
func NumeroTrainValide(numero_train int, nb_dispo int) bool {
	return numero_train >= 1 && numero_train <= nb_dispo
}

func NombrePlaceValide(nb_place int, max int) bool {
	return nb_place > 0 && nb_place <= max
}

func NombrePlaceValideAnnulation(nb_place int, max int, achat int) bool {
	return nb_place > 0 && nb_place <= max && nb_place <= achat
}

// consulter les billets (local)
func ConsulterBillet(numero_train int) (Billet, error) {
	mu.Lock()
	defer mu.Unlock()

	billet, ok := Stock[numero_train]
	if !ok {
		return Billet{}, errors.New("train non trouvé")
	}
	return billet, nil
}

func SendToCtl(sender *string, msg string, debug bool) {
	pipeName := "/tmp/in_C" + (*sender)[1:]

	err := WriteToFile(pipeName, msg)
	if err != nil {
		fmt.Println("Erreur lors de l'écriture dans le pipe:", err)
		return
	}

	if debug {
		Display_w(sender, "sendToCtl", "Envoi de "+msg)
	}
}

// recevoir du msg
func ReceiveFromCtl(p_nom *string, trains map[int]Billet, debug bool) {
	pipeFile, err := os.Open("/tmp/in_" + *p_nom)
	if err != nil {
		fmt.Println("Erreur d'ouverture du pipe:", err)
		return
	}
	defer pipeFile.Close()

	// lire
	pipeScanner := bufio.NewScanner(pipeFile)
	for pipeScanner.Scan() {
		pipeData := pipeScanner.Text()
		if debug {
			Display_w(p_nom, "ReceiveFromCtl", "Reçu: "+pipeData)
		}

		typeMsg := Findval(p_nom, pipeData, "type")

		switch typeMsg {
		case "debutSC":
			//en_section_critique = true

		case "maj_place":
			num_train, err := strconv.Atoi(Findval(p_nom, pipeData, "num_train"))
			if err != nil {
				Display_e(p_nom, "ReceiveFromCtl", "Erreur de conversion du numéro de train "+err.Error())
			}
			nb_places_restantes, err := strconv.Atoi(Findval(p_nom, pipeData, "nb_place_restantes"))
			if err != nil {
				Display_e(p_nom, "ReceiveFromCtl", "Erreur de conversion nmobre de places restantes "+err.Error())
			}
			if tr, ok := trains[num_train]; ok {
				tr.PlacesRest = nb_places_restantes
				trains[num_train] = tr
			} else {
				fmt.Println("Train", num_train, " non trouvé")
			}

		default:
			fmt.Printf("[WARNING] Type de message inconnu : %s\n", typeMsg)
		}

	}

	if err := pipeScanner.Err(); err != nil {
		fmt.Println("Erreur de lecture du pipe:", err)
	}
}

// Fonction d’écriture dans un FIFO nommé
func WriteToFile(pipeName string, msg string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	file, err := os.OpenFile(pipeName, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, msg)
	return err
}
