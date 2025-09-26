package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"sr05/src/utils"
	"strconv"
	"strings"
)

type Couple struct {
	Str     string
	Horloge int
}

func msg_from_app(p_nom *string, rcvmsg string) bool {
	if utils.Findval(p_nom, rcvmsg, "from") == "A"+(*p_nom)[1:] {
		return true
	} else {
		return false
	}
}

func msg_from_net(p_nom *string, rcvmsg string) bool {
	//fmt.Println(utils.Findval(p_nom, rcvmsg, "from") == "N"+(*p_nom)[1:])
	if utils.Findval(p_nom, rcvmsg, "from") == "N"+(*p_nom)[1:] {
		return true
	} else {
		return false
	}
}

func recaler(x, y int) int {
	if x < y {
		return y + 1
	}
	return x + 1
}

func modif_tableau_site(tableau_site map[int]map[int]Couple, train int, site int, str string, horloge int) {
	tableau_site[train][site] = Couple{Str: str, Horloge: horloge}
}

func initialisationCtrl() (*string, int, int, int, int, map[int]map[int]Couple, bool) {
	var h = 0
	var demande_session_critique = true

	p_nom := flag.String("nom", "controle", "nom")
	p_num := flag.String("numero", "0", "numéro de l'application")
	p_nb_site := flag.String("nb_sites", "1", "nombre de sites dans le réseau")
	p_nb_train := flag.String("nb_trains", "1", "nombre de trains")
	flag.Parse()

	//convertir les chaînes de caractères en entiers
	num, _ := strconv.Atoi(*p_num)
	nb_site, _ := strconv.Atoi(*p_nb_site)
	nb_train, _ := strconv.Atoi(*p_nb_train)

	var tableau_site = make(map[int]map[int]Couple)
	for train := 1; train < nb_train+1; train++ {
		tableau_site[train] = make(map[int]Couple)
		for site := 1; site < nb_site+1; site++ {
			tableau_site[train][site] = Couple{"libération", 0}
		}
	}
	return p_nom, num, nb_site, nb_train, h, tableau_site, demande_session_critique

}

func afficheTableauSite(tab map[int]map[int]Couple) string {
	str := ""

	for _, innerMap := range tab {
		for _, couple := range innerMap {
			str += couple.Str + " " + strconv.Itoa(couple.Horloge) + " "
		}
		str += ","
	}

	return str
}

func afficheCleTableauSite(tab map[int]map[int]Couple) string {
	str := ""
	keys := make([]int, 0, len(tab))
	for key := range tab[1] {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	for i := 0; i < len(keys); i++ {
		str += strconv.Itoa(keys[i]) + " "
	}
	str = str[:len(str)-1]
	return str
}

func parseTabSiteFromString(input string, keyStr string) map[int]map[int]Couple {
	result := make(map[int]map[int]Couple)
	keys := strings.Fields(keyStr)

	groups := strings.Split(input, ",")
	for outerIndex, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}

		tokens := strings.Fields(group)
		inner := make(map[int]Couple)
		iterKeys := 0
		key := 0

		for i := 0; i < len(tokens)-1; i += 2 {
			text := tokens[i]
			horloge, _ := strconv.Atoi(tokens[i+1])
			key, _ = strconv.Atoi(keys[iterKeys])
			inner[key] = Couple{Str: text, Horloge: horloge}
			iterKeys++
		}

		result[outerIndex+1] = inner
	}

	return result
}

func initHorlogeVectorielle(hlg_vect map[int]int) {
	for i := 0; i < len(hlg_vect); i++ {
		hlg_vect[i] = 0
	}
}

func afficheHorlogeVectorielle(hlg_vect map[int]int) string {
	keys := make([]int, 0, len(hlg_vect))
	for key := range hlg_vect {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	str := strconv.Itoa(hlg_vect[keys[0]])
	for i := 1; i < len(hlg_vect); i++ {
		str += " " + strconv.Itoa(hlg_vect[keys[i]])
	}
	return str
}

func ensureVectorSize(hlg_vect *[]int, requiredLen int) {
	for len(*hlg_vect) < requiredLen {
		*hlg_vect = append(*hlg_vect, 0)
	}
}

func recalerHorlogeVectorielle(hlg_vect_local map[int]int, hlg_vect_externe []int) {
	str := "test"
	utils.Display_w(&str, "hlg_vect_externe : ", fmt.Sprintf("%v %v", hlg_vect_externe, hlg_vect_local))

	keys := make([]int, 0, len(hlg_vect_local))
	for key := range hlg_vect_local {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	for i := 0; i < len(keys); i++ {
		hlg_vect_local[keys[i]] = max(hlg_vect_local[keys[i]], hlg_vect_externe[i])
	}
}

func recupererHorlogeVectorielle(hlg_in_msg string) []int {
	champs := strings.Fields(hlg_in_msg)

	var tableau []int
	for i := 0; i < len(champs); i++ {
		val, err := strconv.Atoi(champs[i])
		if err != nil {
			fmt.Println("Erreur de conversion :", err)
			continue
		}
		tableau = append(tableau, val)
	}
	return tableau
}

func main() {

	p_nom, p_num, nb_site, nb_train, h, tableau_site, demande_session_critique := initialisationCtrl()
	debug := true

	couleur := "blanc"
	bilan := 0
	nbEtatsAttendus := 0
	nbMsgAttendus := 0
	nbVoisins := 0
	etat_local := ""
	initiateur := false
	hlg_vect := make(map[int]int, nb_site)
	initHorlogeVectorielle(hlg_vect)

	section_critique_en_cours := false

	utils.Display_d(p_nom, "main", "initialisation du controlleur: "+*p_nom+" "+strconv.Itoa(p_num)+" "+strconv.Itoa(nb_site)+" "+strconv.Itoa(nb_train)+" "+strconv.Itoa(h))

	reader := bufio.NewReader(os.Stdin)
	for {

		rcvmsg, err := reader.ReadString('\n')
		rcvmsg = strings.TrimSpace(rcvmsg)

		if rcvmsg != "" {
			utils.Msg_receive(p_nom, rcvmsg)
		}

		if err != nil {
			if debug {
				utils.Display_e(p_nom, "main", "erreur de lecture"+err.Error())
			}
			continue
		}

		num_train, _ := strconv.Atoi(utils.Findval(p_nom, rcvmsg, "num_train"))

		// traitement du message
		var type_msg = utils.Findval(p_nom, rcvmsg, "type")

		if msg_from_app(p_nom, rcvmsg) == true {
			if debug {
				utils.Display_e(p_nom, "main", "message provenant de l'application: "+type_msg)
			}
			h = h + 1
			hlg_vect[p_num]++

			switch type_msg {
			case "demandeSC":
				hlg_vect[p_num]++
				bilan += nb_site - 1
				modif_tableau_site(tableau_site, num_train, p_num, "requete", h)
				utils.Display_d(p_nom, "main", "demandeSC - tableau site : "+fmt.Sprint(tableau_site[num_train]))
				if debug {
					utils.Display_e(p_nom, "main", "demandeSC - tableau site : "+fmt.Sprint(tableau_site))
				}
				utils.Msg_send(p_nom, "requete "+strconv.Itoa(h))
				fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("num_train", strconv.Itoa(num_train)) + utils.Msg_format("type", "requete") + utils.Msg_format("hlg", strconv.Itoa(h)) + utils.Msg_format("couleur_msg", couleur) + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))

			case "finSC":
				section_critique_en_cours = false
				hlg_vect[p_num]++
				bilan += nb_site - 1
				nb_place_restantes_str := utils.Findval(p_nom, rcvmsg, "nb_place_restantes")
				modif_tableau_site(tableau_site, num_train, p_num, "liberation", h)
				utils.Display_d(p_nom, "main", "finSC - tableau site : "+fmt.Sprint(tableau_site[num_train]))
				if debug {
					utils.Display_e(p_nom, "main", "finSC - tableau site : "+fmt.Sprint(tableau_site))
				}
				utils.Msg_send(p_nom, "liberation "+strconv.Itoa(h))
				fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("num_train", strconv.Itoa(num_train)) + utils.Msg_format("nb_place_restantes", nb_place_restantes_str) + utils.Msg_format("type", "liberation") + utils.Msg_format("hlg", strconv.Itoa(h)+utils.Msg_format("couleur_msg", couleur)+utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect))))

			case "reponseEtatLocalApp":
				hlg_vect[p_num]++
				etat_local = utils.Findval(p_nom, rcvmsg, "etatLocal") + " date : " + afficheHorlogeVectorielle(hlg_vect)
				couleur = "rouge"
				if !initiateur {
					fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("bilan", strconv.Itoa(bilan)) + utils.Msg_format("type", "etat") + utils.Msg_format("etatLocal", etat_local) + utils.Msg_format("couleur_msg", couleur) + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
				} else {
					fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("type", "debutSnapshot") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))

				}
			case "initSnapshot":
				initiateur = true
				nbEtatsAttendus = nb_site - 1
				nbMsgAttendus = bilan
				hlg_vect[p_num]++
				fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("type", "demandeEtatLocal") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
			case "demande_quit_app":
				utils.Msg_send(p_nom, "demande_quit_net")
				hlg_vect[p_num]++
				fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "demande_quit_net"))
			case "reponseTabTrains":
				hlg_vect[p_num]++
				fmt.Println(utils.Msg_format("type", "tabTrainsToNewSite") + utils.Msg_format("tabTrains", utils.Findval(p_nom, rcvmsg, "tabTrains")) + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
			}
		} else if msg_from_net(p_nom, rcvmsg) == true {
			if debug {
				utils.Display_e(p_nom, "main", "message provenant du net: "+type_msg)
			}

			switch type_msg {
			case "quit_ok_net":
				utils.Msg_send(p_nom, "quit_ok_app")
				fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("type", "quit_ok_app"))
				os.Exit(0)
			}

		} else {

			//si le message recu est à destination de l'app on ne le traite pas
			if type_msg != "liberation" && type_msg != "requete" && type_msg != "accusé" && type_msg != "etat" && type_msg != "prepost" && type_msg != "debutSnapshot" && type_msg != "finSnapshot" && type_msg != "nouveau_site" && type_msg != "replicat" && type_msg != "tabTrainsToNewSite" && type_msg != "departAutreSite" && type_msg != "newNbVoisins" {
				continue
			}

			h_receive, _ := strconv.Atoi(utils.Findval(p_nom, rcvmsg, "hlg"))
			if err != nil {
				if debug {
					utils.Display_e(p_nom, "main", "erreur de conversion"+err.Error())
				}
				//	continue
			}

			var from_to = utils.Findval(p_nom, rcvmsg, "from")
			from_num, _ := strconv.Atoi(utils.Findval(p_nom, rcvmsg, "from_num"))
			if err != nil {
				if debug {
					utils.Display_e(p_nom, "main", "erreur de conversion"+err.Error())
				}
				//	continue
			}

			if utils.Findval(p_nom, rcvmsg, "to") != "" {
				if utils.Findval(p_nom, rcvmsg, "to") != *p_nom {
					if debug {
						utils.Display_e(p_nom, "main", "message non destinataire")
					}
					continue
				}
			}

			if type_msg != "nouveau_site" && type_msg != "newNbVoisins" && type_msg != "departAutreSite" {
				hlg_vect_extern := recupererHorlogeVectorielle(utils.Findval(p_nom, rcvmsg, "hlg_vect"))
				recalerHorlogeVectorielle(hlg_vect, hlg_vect_extern)
				hlg_vect[p_num]++
			} else {
				hlg_vect[p_num]++ // Action interne
			}

			switch type_msg {
			case "liberation": //plus faire mise a jour du nombre de place restantes
				//######## Snapshot :
				bilan--
				couleurMsg := utils.Findval(p_nom, rcvmsg, "couleur_msg")
				if couleurMsg == "rouge" && couleur == "blanc" {
					hlg_vect[p_num]++
					fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("type", "demandeEtatLocal") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
				} else if couleurMsg == "blanc" && couleur == "rouge" {
					hlg_vect[p_num]++
					if initiateur {
						nbMsgAttendus-- // Sinon l'initiateur ne comptera pas avoir reçu un prepost
						etat_local += "\n" + utils.Msg_format("type", "prepost") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)) + rcvmsg + utils.Msg_format("msg_to", *p_nom)
					} else {
						fmt.Println(utils.Msg_format("type", "prepost") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)) + rcvmsg + utils.Msg_format("msg_to", *p_nom))
					}
				}

				//traiter le message
				//######## Mutex :
				h = recaler(h, h_receive)
				modif_tableau_site(tableau_site, num_train, from_num, "liberation", h_receive)
				utils.Display_d(p_nom, "main", "liberation - tableau site : "+fmt.Sprint(tableau_site[num_train]))
				if debug {
					utils.Display_e(p_nom, "main", "liberation - tableau site : "+fmt.Sprint(tableau_site))
				}

				//mettre à jour le nombre de places restantes
				nb_place_restantes_maj, err := strconv.Atoi(utils.Findval(p_nom, rcvmsg, "nb_place_restantes"))
				if err != nil {
					if debug {
						utils.Display_e(p_nom, "main", "erreur de conversion du nombre de place restantes"+err.Error())
					}
					continue
				}
				utils.Msg_send(p_nom, "nb_place_restantes "+strconv.Itoa(nb_place_restantes_maj))
				hlg_vect[p_num]++
				fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("num_train", strconv.Itoa(num_train)) + utils.Msg_format("nb_place_restantes", strconv.Itoa(nb_place_restantes_maj)) + utils.Msg_format("type", "maj_place") + utils.Msg_format("hlg", strconv.Itoa(h)) + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))

				if tableau_site[num_train][p_num].Str == "requete" {
					demande_session_critique = true
					//for k := 1; k < len(tableau_site[num_train])+1; k++ {
					for k, _ := range tableau_site[num_train] {
						if k != p_num {
							if tableau_site[num_train][p_num].Horloge > tableau_site[num_train][k].Horloge || (tableau_site[num_train][p_num].Horloge == tableau_site[num_train][k].Horloge && p_num > k) {
								demande_session_critique = false
								break
							}
						}
					}

					if demande_session_critique {
						utils.Msg_send(p_nom, "débutSC "+strconv.Itoa(h))
						//envoyer à l'app de base
						hlg_vect[p_num]++
						fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("type", "débutSC") + utils.Msg_format("hlg", strconv.Itoa(h)) + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
					}
				}

			case "requete":
				//######## Snapshot :
				bilan--
				couleurMsg := utils.Findval(p_nom, rcvmsg, "couleur_msg")
				if couleurMsg == "rouge" && couleur == "blanc" {
					hlg_vect[p_num]++
					fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("type", "demandeEtatLocal") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
				} else if couleurMsg == "blanc" && couleur == "rouge" {
					hlg_vect[p_num]++
					if initiateur {
						nbMsgAttendus-- // Sinon l'initiateur ne comptera pas avoir reçu un prepost
						etat_local += "\n" + utils.Msg_format("type", "prepost") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)) + rcvmsg + utils.Msg_format("msg_to", *p_nom)
					} else {
						fmt.Println(utils.Msg_format("type", "prepost") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)) + rcvmsg + utils.Msg_format("msg_to", *p_nom))
					}
				}

				//######## Mutex :
				h = recaler(h, h_receive)
				modif_tableau_site(tableau_site, num_train, from_num, "requete", h_receive)

				utils.Display_d(p_nom, "main", "requete - tableau site : "+fmt.Sprint(tableau_site[num_train]))
				if debug {
					utils.Display_e(p_nom, "main", "requete - tableau site : "+fmt.Sprint(tableau_site))
				}
				utils.Msg_send(p_nom, "accusé "+strconv.Itoa(h))
				hlg_vect[p_num]++
				bilan++

				fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("to", from_to) + utils.Msg_format("num_train", strconv.Itoa(num_train)) + utils.Msg_format("type", "accusé") + utils.Msg_format("hlg", strconv.Itoa(h)) + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))

				if tableau_site[num_train][p_num].Str == "requete" && !section_critique_en_cours {
					demande_session_critique = true
					for k, _ := range tableau_site[num_train] {
						if k != p_num {
							if tableau_site[num_train][p_num].Horloge > tableau_site[num_train][k].Horloge || (tableau_site[num_train][p_num].Horloge == tableau_site[num_train][k].Horloge && p_num > k) {
								demande_session_critique = false
								break
							}
						}
					}

					if demande_session_critique {
						section_critique_en_cours = true
						utils.Msg_send(p_nom, "débutSC "+strconv.Itoa(h))
						//envoyer à l'app de base
						hlg_vect[p_num]++
						fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("num_train", strconv.Itoa(num_train)) + utils.Msg_format("type", "débutSC") + utils.Msg_format("hlg", strconv.Itoa(h)) + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
					}
				}

			case "accusé":
				//######## Snapshot :
				bilan--
				couleurMsg := utils.Findval(p_nom, rcvmsg, "couleur_msg")
				if couleurMsg == "rouge" && couleur == "blanc" {
					hlg_vect[p_num]++
					fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("type", "demandeEtatLocal") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
				} else if couleurMsg == "blanc" && couleur == "rouge" {
					hlg_vect[p_num]++
					if initiateur {
						nbMsgAttendus-- // Sinon l'initiateur ne comptera pas avoir reçu un prepost
						etat_local += "\n" + utils.Msg_format("type", "prepost") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)) + rcvmsg + utils.Msg_format("msg_to", *p_nom)
					} else {
						fmt.Println(utils.Msg_format("type", "prepost") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)) + rcvmsg + utils.Msg_format("msg_to", *p_nom))
					}
				}

				//######## Mutex :
				h = recaler(h, h_receive)
				//regarder si le message m'est destiné (champ to)
				if tableau_site[num_train][from_num].Str != "requete" {
					modif_tableau_site(tableau_site, num_train, from_num, "accusé", h_receive)
				}
				utils.Display_d(p_nom, "main", "accusé - tableau site : "+fmt.Sprint(tableau_site[num_train]))
				if debug {
					utils.Display_e(p_nom, "main", "accusé - tableau site : "+fmt.Sprint(tableau_site))
				}

				if tableau_site[num_train][p_num].Str == "requete" {
					demande_session_critique = true
					for k, _ := range tableau_site[num_train] {
						if k != p_num {
							if tableau_site[num_train][p_num].Horloge > tableau_site[num_train][k].Horloge || (tableau_site[num_train][p_num].Horloge == tableau_site[num_train][k].Horloge && p_num > k) {
								demande_session_critique = false
								break
							}
						}
					}

					if demande_session_critique {
						section_critique_en_cours = true
						utils.Msg_send(p_nom, "débutSC "+strconv.Itoa(h))
						//envoyer à l'app de base
						hlg_vect[p_num]++
						fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("num_train", strconv.Itoa(num_train)) + utils.Msg_format("type", "débutSC") + utils.Msg_format("hlg", strconv.Itoa(h)) + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
					}
				}
			case "etat":
				if initiateur {
					etat_local += "\n" + utils.Findval(p_nom, rcvmsg, "etatLocal")
					nbEtatsAttendus--
					bilanTmp, _ := strconv.Atoi(utils.Findval(p_nom, rcvmsg, "bilan"))
					nbMsgAttendus += bilanTmp

					utils.Display_w(p_nom, "nbMessageAttendus : ", strconv.Itoa(nbMsgAttendus))
					utils.Display_w(p_nom, "nbVoisins : ", strconv.Itoa(nbVoisins))

					if nbEtatsAttendus == 0 && nbMsgAttendus == 0 {
						hlg_vect[p_num]++
						fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("type", "finSnapshot") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
						initiateur = false
						couleur = "blanc"
						utils.Display_w(p_nom, "Snapshot : ", etat_local)
						fichier, err := os.OpenFile("Snapshot_Site"+strconv.Itoa(p_num)+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						if err != nil {
							fmt.Println("Erreur d'ouverture :", err)
							return
						}
						defer fichier.Close()

						// Écriture dans le fichier
						fichier.WriteString(etat_local + "\n\n\n")
					}
				}
			case "prepost":
				if initiateur {
					nbMsgAttendus--
					etat_local += "\n" + rcvmsg

					utils.Display_w(p_nom, "nbMessageAttendus : ", strconv.Itoa(nbMsgAttendus))
					utils.Display_w(p_nom, "nbVoisins : ", strconv.Itoa(nbVoisins))

					if nbEtatsAttendus == 0 && nbMsgAttendus == 0 {
						hlg_vect[p_num]++
						fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("type", "finSnapshot") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
						initiateur = false
						couleur = "blanc"
						utils.Display_w(p_nom, "Snapshot : ", etat_local)
						fichier, err := os.OpenFile("Snapshot_Site"+strconv.Itoa(p_num)+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						if err != nil {
							fmt.Println("Erreur d'ouverture :", err)
							return
						}
						defer fichier.Close()

						// Écriture dans le fichier
						fichier.WriteString(etat_local + "\n\n\n")
					}
				}

			case "debutSnapshot":
				hlg_vect[p_num]++
				fmt.Println(utils.Msg_format("from", *p_nom) + utils.Msg_format("from_num", strconv.Itoa(p_num)) + utils.Msg_format("type", "demandeEtatLocal") + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
			case "finSnapshot":
				couleur = "blanc"

			case "nouveau_site":
				//un nouveau site a été ajouté dans le réseau --> mettre à jour le tableau local
				new_num_site, _ := strconv.Atoi(utils.Findval(p_nom, rcvmsg, "new_numsite"))
				for train := 1; train < nb_train+1; train++ {
					tableau_site[train][new_num_site] = Couple{"libération", 0}
				}
				nb_site++
				hlg_vect[new_num_site] = 0
				if utils.Findval(p_nom, rcvmsg, "isFather") == "true" {
					nbVoisins++
					hlg_vect[p_num]++
					fmt.Println(utils.Msg_format("type", "replicat") + utils.Msg_format("nbSite", strconv.Itoa(nb_site)) + utils.Msg_format("tableau_site", afficheTableauSite(tableau_site)) + utils.Msg_format("listKeys", afficheCleTableauSite(tableau_site)) + utils.Msg_format("hlg_vect", afficheHorlogeVectorielle(hlg_vect)))
					hlg_vect[p_num]++
					fmt.Println(utils.Msg_format("type", "demandeTabTrains")) // Demander à l'App d'envoyer l'etat des trains
				}
			case "tabTrainsToNewSite":
				hlg_vect[p_num]++
				fmt.Println(utils.Msg_format("type", "updateTrainsToNewSite") + utils.Msg_format("tabTrains", utils.Findval(p_nom, rcvmsg, "tabTrains")))
			case "replicat":
				nbVoisins++
				nb_site, _ = strconv.Atoi(utils.Findval(p_nom, rcvmsg, "nbSite"))
				tableau_site = parseTabSiteFromString(utils.Findval(p_nom, rcvmsg, "tableau_site"), utils.Findval(p_nom, rcvmsg, "listKeys")) // Le nouveau site doit avoir le replicat
				for i := 1; i <= nb_site; i++ {
					hlg_vect[i] = 0
				}
			case "departAutreSite":
				nb_site--
				numsiteDepart, _ := strconv.Atoi(utils.Findval(p_nom, rcvmsg, "numsiteDepart"))
				for train := 1; train < nb_train+1; train++ {
					delete(tableau_site[train], numsiteDepart)
				}
				delete(hlg_vect, numsiteDepart)
				utils.Display_w(p_nom, "hlg_vect: ", fmt.Sprintf("%v", hlg_vect))
			case "newNbVoisins":
				nbVoisins, _ = strconv.Atoi(utils.Findval(p_nom, rcvmsg, "nbVoisins"))
			}

		}
	}
}
