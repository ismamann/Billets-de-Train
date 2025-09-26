package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sr05/src/utils"
	"strconv"
	"strings"
	"sync"
)

var (
	en_session_critique = false
	mutexApp            = &sync.Mutex{}
	ch                  = make(chan bool)
)

func initialisationApp() (*string, bool) {
	p_nom := flag.String("nom", "app", "nom")
	debugStr := flag.String("debug", "false", "debug")
	flag.Parse()

	debug, err := strconv.ParseBool(*debugStr)
	if err != nil {
		fmt.Println("Erreur lors de la conversion de debug:", err)
		return nil, false
	}

	return p_nom, debug
}

func afficheTableauTrain(trains map[int]utils.Billet) string {
	str := strconv.Itoa(trains[1].PlacesRest)
	for i := 2; i <= len(trains); i++ {
		str += " " + strconv.Itoa(trains[i].PlacesRest)
	}
	return str
}

func parseTableauTrainFromString(str string, trains map[int]utils.Billet) {
	champs := strings.Fields(str)
	var tr utils.Billet
	for i := 0; i < len(champs); i++ {
		tr = trains[i+1]
		tr.PlacesRest, _ = strconv.Atoi(champs[i])
		trains[i+1] = tr
	}

}

func ReceiveFromCtl(p_nom *string, trains map[int]utils.Billet, debug bool) {
	for {
		pipeFile, err := os.Open("/tmp/in_" + *p_nom)
		if err != nil {
			fmt.Println("Erreur d'ouverture du pipe:", err)
			return
		}
		defer pipeFile.Close()

		// lire
		pipeScanner := bufio.NewScanner(pipeFile)
		for pipeScanner.Scan() {
			mutexApp.Lock()
			pipeData := pipeScanner.Text()
			if debug {
				utils.Display_w(p_nom, "ReceiveFromCtl", "Reçu: "+pipeData)
			}

			typeMsg := utils.Findval(p_nom, pipeData, "type")
			switch typeMsg {
			case "débutSC":
				en_session_critique = true
				ch <- en_session_critique

			case "maj_place":
				num_train, err := strconv.Atoi(utils.Findval(p_nom, pipeData, "num_train"))
				if err != nil {
					utils.Display_e(p_nom, "ReceiveFromCtl", "Erreur de conversion du numéro de train "+err.Error())
				}
				nb_places_restantes, err := strconv.Atoi(utils.Findval(p_nom, pipeData, "nb_place_restantes"))
				if err != nil {
					utils.Display_e(p_nom, "ReceiveFromCtl", "Erreur de conversion nmobre de places restantes "+err.Error())
				}
				if tr, ok := trains[num_train]; ok {
					tr.PlacesRest = nb_places_restantes
					trains[num_train] = tr
				} else {
					fmt.Println("Train", num_train, " non trouvé")
				}
			case "demandeEtatLocal":
				msg := fmt.Sprintf(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "reponseEtatLocalApp") + utils.Msg_format("etatLocal", utils.AfficherTrainEtatLocal(trains, p_nom)))
				utils.SendToCtl(p_nom, msg, debug)
			case "quit_ok_app":
				utils.AfficherArret()
				os.Exit(0)
			case "demandeTabTrains":
				utils.SendToCtl(p_nom, utils.Msg_format("type", "reponseTabTrains")+utils.Msg_format("tabTrains", afficheTableauTrain(trains))+utils.Msg_format("from", *p_nom), debug)
			case "updateTrainsToNewSite":
				parseTableauTrainFromString(utils.Findval(p_nom, pipeData, "tabTrains"), trains)
			default:
				if debug {
					utils.Display_w(p_nom, "ReceiveFromCtl", "Message de type inconnu")
				}
			}
			mutexApp.Unlock()

		}

		if err := pipeScanner.Err(); err != nil {
			fmt.Println("Erreur de lecture du pipe:", err)
		}
	}
}

func main() {

	p_nom, debug := initialisationApp()

	trains := utils.InitStock()
	achat := utils.InitAchat()

	go ReceiveFromCtl(p_nom, trains, debug)

	fmt.Println(utils.Underline + utils.BrightBlue + "Bienvenue sur l'application d'achat de billet de train." + utils.Raz + utils.Raz)
	for {
		utils.AfficherTrain(trains)
		utils.AfficherMenu()

		choix, err := utils.ChoixUser()
		mutexApp.Lock()
		if err != nil {
			fmt.Println("Erreur lors de la saisie du choix:", err)
			mutexApp.Unlock()
			continue
		}

		switch choix {
		case 0:
			//quitter
			fmt.Println("Vous avez choisi de quitter l'application.")
			msg := fmt.Sprintf(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "demande_quit_app"))
			utils.SendToCtl(p_nom, msg, debug)
			//peut etre faire le meme système que pour la session critique en attendant qu'une variable se mette à true
		case 1:
			//achat d'un billet
			fmt.Println("Vous avez choisi d'acheter un billet.")
			utils.AfficherTrain(trains)
			fmt.Println("\nSi vous vous êtes trompé de choix, entrez " + utils.Bold + "0." + utils.Raz)
			numero_train, err := utils.GetChoixUser(utils.Underline + "Entrez le numero du train : \n" + utils.Raz)
			if utils.AnnulerChoix(numero_train) {
				break
			}

			for err != nil || !utils.NumeroTrainValide(numero_train, len((trains))) {
				fmt.Println("Numero de train invalide. Veuillez entrer un choix entre 1 et", len(trains), ", "+utils.Bold+"0"+utils.Raz+" pour abandonner.\n")
				numero_train, err = utils.GetChoixUser(utils.Underline + "Entrez le numero du train : \n" + utils.Raz)
				if utils.AnnulerChoix(numero_train) {
					break
				}
			}

			//demande de section critique
			msg := fmt.Sprintf(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "demandeSC") + utils.Msg_format("num_train", strconv.Itoa(numero_train)))
			utils.SendToCtl(p_nom, msg, debug)

			mutexApp.Unlock()

			en_session_critique = <-ch

			mutexApp.Lock()

			fmt.Println("\n" + utils.Bold + "0" + utils.Raz + " pour abandonner.")
			nb_place, err := utils.GetChoixUser(utils.Underline + "Combien de places voulez-vous acheter : \n" + utils.Raz)
			if utils.AnnulerChoix(nb_place) {
				msg := fmt.Sprintf(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "finSC") + utils.Msg_format("nb_place_restantes", strconv.Itoa(trains[numero_train].PlacesRest)) + utils.Msg_format("num_train", strconv.Itoa(numero_train)))
				utils.SendToCtl(p_nom, msg, debug)
				en_session_critique = false
				break
			}

			for err != nil || !utils.NombrePlaceValide(nb_place, trains[numero_train].PlacesRest) {
				fmt.Println("Nombre de places invalide ou Nombre de places demandées supérieur au nombre de places disponibles.\n")
				nb_place, err = utils.GetChoixUser(utils.Underline + "Combien de places voulez-vous acheter : \n" + utils.Raz)
				if utils.AnnulerChoix(nb_place) {
					break
				}
			}

			//on peut modifier les places restantes
			if tr, ok := trains[numero_train]; ok {
				tr.PlacesRest -= nb_place
				trains[numero_train] = tr
			} else {
				fmt.Println("train", numero_train, "non trouvé.")
				break
			}

			if ach, ok := achat[numero_train]; ok {
				ach.Nb += nb_place
				achat[numero_train] = ach
			} else {
				fmt.Println("train", numero_train, "non trouvé dans les achats de billets.")
				break
			}

			msg = fmt.Sprintf(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "finSC") + utils.Msg_format("nb_place_restantes", strconv.Itoa(trains[numero_train].PlacesRest)) + utils.Msg_format("num_train", strconv.Itoa(numero_train)))
			utils.SendToCtl(p_nom, msg, debug)
			fmt.Printf("Vous avez acheté %d places pour le train numéro %d.\n\n", nb_place, numero_train)
			en_session_critique = false

		case 2:
			//annulation d'un billet

			fmt.Println("Vous avez choisi d'anuler un billet.")
			utils.AfficherTrain(trains)
			fmt.Println("\nSi vous vous êtes trompé de choix, entrez " + utils.Bold + "0." + utils.Raz)
			numero_train, err := utils.GetChoixUser(utils.Underline + "Entrez le numero du train : \n" + utils.Raz)
			if utils.AnnulerChoix(numero_train) {
				break
			}

			for err != nil || !utils.NumeroTrainValide(numero_train, len((trains))) {
				fmt.Println("Numero de train invalide. Veuillez entrer un choix entre 1 et", len(trains), ", "+utils.Bold+"0"+utils.Raz+" pour abandonner.\n")
				numero_train, err = utils.GetChoixUser(utils.Underline + "Entrez le numero du train : \n" + utils.Raz)
				if utils.AnnulerChoix(numero_train) {
					break
				}
			}

			//demande de section critique
			msg := fmt.Sprintf(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "demandeSC") + utils.Msg_format("num_train", strconv.Itoa(numero_train)))
			utils.SendToCtl(p_nom, msg, debug)

			mutexApp.Unlock()

			en_session_critique = <-ch

			mutexApp.Lock()

			fmt.Println("\n" + utils.Bold + "0" + utils.Raz + " pour abandonner.")
			nb_place, err := utils.GetChoixUser(utils.Underline + "Combien de places voulez-vous remettre : \n" + utils.Raz)
			if utils.AnnulerChoix(nb_place) {
				msg := fmt.Sprintf(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "finSC") + utils.Msg_format("nb_place_restantes", strconv.Itoa(trains[numero_train].PlacesRest)) + utils.Msg_format("num_train", strconv.Itoa(numero_train)))
				utils.SendToCtl(p_nom, msg, debug)
				en_session_critique = false
				break
			}

			for err != nil || !utils.NombrePlaceValideAnnulation(nb_place, trains[numero_train].CapaciteMax-trains[numero_train].PlacesRest, achat[numero_train].Nb) {
				fmt.Println("Nombre de places invalide ou Nombre de places supérieur au nombre de places total du train ou Nombre de places demandées supérieur au nombre de places achetées.")
				nb_place, err = utils.GetChoixUser(utils.Underline + "Combien de paces voulez-vous remettre : \n" + utils.Raz)
				if utils.AnnulerChoix(nb_place) {
					break
				}
			}

			//on peut modifier les places restantes
			if tr, ok := trains[numero_train]; ok {
				tr.PlacesRest += nb_place
				trains[numero_train] = tr
			} else {
				fmt.Println("train", numero_train, "non trouvé.")
				break
			}

			if ach, ok := achat[numero_train]; ok {
				ach.Nb -= nb_place
				achat[numero_train] = ach
			} else {
				fmt.Println("train", numero_train, "non trouvé dans les achats de billets.")
				break
			}

			msg = fmt.Sprintf(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "finSC") + utils.Msg_format("nb_place_restantes", strconv.Itoa(trains[numero_train].PlacesRest)) + utils.Msg_format("num_train", strconv.Itoa(numero_train)))
			utils.SendToCtl(p_nom, msg, debug)
			fmt.Printf("Vous avez relaché %d places pour le train numéro %d.\n\n", nb_place, numero_train)
			en_session_critique = false

		case 3:
			//sauvegarde des données
			fmt.Println("Vous avez choisi d'effectuer une sauvegarde.")
			fmt.Printf("Nom de la sauvegarde local: %sSnapshot_Site%s.txt%s .\n\n", utils.Bold, (*p_nom)[1:], utils.Raz)
			msg := fmt.Sprintf(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "initSnapshot"))
			utils.SendToCtl(p_nom, msg, debug)
		}
		mutexApp.Unlock()

	}

}
