package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"sr05/src/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/perlin-network/noise"
)

var debug = false

var mutex = &sync.Mutex{}
var chNet = make(chan int, 200)
var parent = ""
var nbVoisinsAttendus = 0
var demandeDepart = false
var p_nom *string
var p_num = ""
var listVagues = make(map[string]Vague)
var p_nom_demandeur string
var reader = bufio.NewReader(os.Stdin)
var tableauCopieTermine = 2

var fieldsep = "/"
var keyvalsep = "="

var fieldsep2 = "!"
var keyvalsep2 = ";"

func Msg_format(key string, val string) string {
	return fieldsep + keyvalsep + key + keyvalsep + val
}

func Msg_format2(key string, val string) string {
	return fieldsep2 + keyvalsep2 + key + keyvalsep2 + val
}

func Findval(msg string, key string) string {
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)
	for _, keyval := range tab_allkeyvals {
		equ := keyval[0:1]
		tabkeyval := strings.Split(keyval[1:], equ)
		if tabkeyval[0] == key {
			return tabkeyval[1]
		}
	}
	return ""
}

// Vérifie les erreurs et panique en cas d'erreur
func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Election struct {
	NbVoisinsAtt int
	Elu          int
	Parent       string
	Demandeur    noise.HandlerContext
}

type Vague struct {
	NbVoisinsAttendus int
	Parent            string
}

// Structure de l'application
type App struct {
	p_nom               string
	name                string
	pere                string
	fils                []string
	voisins             []string
	node                *noise.Node
	accepted_in_network bool
	election            *Election
}

// Filtre les voisins en excluant un voisin spécifique
func filtrerVoisins(voisins []string, exclude string) []string {
	var filtered []string
	for _, voisin := range voisins {
		if voisin != exclude {
			filtered = append(filtered, voisin)
		}
	}
	return filtered
}

// Affiche voisins
func (self *App) afficheVoisins() string {
	//si self.fils est vide on affiche un message approprié
	if len(self.fils) == 0 {
		return "Aucun fils"
	}
	var str = self.fils[0]
	for i := 1; i < len(self.fils); i++ {
		if self.fils[i] != self.pere {
			str += " " + self.fils[i]
		}
	}
	return str
}

// Récuperer voisins
func (self *App) recupererVoisins(list_voisins string) {
	if list_voisins != "Aucun fils" {
		champs := strings.Fields(list_voisins)

		for _, champ := range champs {
			if champ != self.node.ID().Address { // Cas où la racine part et la nouvelle racine est dans la liste de voisin reçu
				self.voisins = append(self.voisins, champ)
				self.fils = append(self.fils, champ)
			}
		}
	}
}

// Envoie un message à plusieurs cibles
func (self *App) send_msg_targets(msg string, targets []string) {
	utils.Msg_send(&self.p_nom, msg)
	for _, target := range targets {
		self.send_msg(target, msg)
	}
}

func (self *App) send_msg(cible string, msg string) {
	err := self.node.Send(context.TODO(), cible, []byte(msg))
	check(err)
}

// Envoie une demande d'admission
func (self *App) requete_admission(cible string) {
	utils.Msg_send(&self.p_nom, "demande_admission")
	msg := Msg_format("type", "demande_admission") + Msg_format("nom", *p_nom)
	self.send_msg(cible, msg)
}

// Accepte une demande d'admission
func (self *App) accept_admission(ctx noise.HandlerContext) {
	utils.Display_d(&self.p_nom, "accept_admission", "Acceptation de la demande d'admission")
	utils.Msg_send(&self.p_nom, "ok_admission")
	msg_type := Msg_format("type", "ok_admission")
	self.send_msg(ctx.ID().Address, msg_type)
	utils.Display_d(&self.p_nom, "accept_admission", "Ajout aux voisins")
	self.voisins = append(self.voisins, ctx.ID().Address)
	self.fils = append(self.fils, ctx.ID().Address)
	nbVoisinsAttendus++
}

// Rejoint le réseau
func (self *App) join_network(cible string) {
	self.voisins = append(self.voisins, cible)
	self.pere = cible
	nbVoisinsAttendus++
}

// Affiche la liste des voisins
func (self *App) list_voisins() {
	fmt.Println("Voisins:")
	for _, voisin := range self.voisins {
		fmt.Println(voisin)
	}
	fmt.Println("Fils:")
	for _, fils := range self.fils {
		fmt.Println(fils)
	}
}

func recoverMessageFromCtrl(msg string) string {
	key := "messageToCtrl="

	parts := strings.Split(msg, key)

	if len(parts) >= 2 {
		return parts[1]
	}

	return ""
}

func (self *App) receive_msg_from_ctrl() {
	var msg = ""
	for {

		rcvmsg, err := reader.ReadString('\n')
		rcvmsg = strings.TrimSpace(rcvmsg)
		mutex.Lock()

		if rcvmsg != "" {
			utils.Msg_receive(p_nom, rcvmsg)
		}

		if err != nil {
			if debug {
				utils.Display_e(p_nom, "main", "erreur de lecture"+err.Error())
			}
			continue
		}

		var type_msg = utils.Findval(p_nom, rcvmsg, "type")

		//si le message recu est à destination de l'app on ne le traite pas
		if type_msg != "liberation" && type_msg != "requete" && type_msg != "accusé" && type_msg != "etat" && type_msg != "prepost" && type_msg != "debutSnapshot" && type_msg != "finSnapshot" && type_msg != "replicat" && type_msg != "tabTrainsToNewSite" && type_msg != "demande_quit_net" {
			mutex.Unlock()
			continue
		}

		if type_msg == "demande_quit_net" {
			nbVoisinsAttendus = len(self.voisins)
			if nbVoisinsAttendus == 0 {
				//Envoyer message au controleur (puis app) pour qu'ils se quittent
				utils.Msg_send(&self.p_nom, "controleur_quitter")
				fmt.Println(utils.Msg_format("from", self.name) + utils.Msg_format("type", "quit_ok_net"))
				os.Exit(0)
			}
			self.initElection()
			self.election.Elu, _ = strconv.Atoi(self.node.Addr()[1:])
			self.election.Parent = self.node.Addr()[1:]
			utils.Msg_send(&self.p_nom, "demande_départ")
			demandeDepart = true
			msg := Msg_format("type", "bleu") + Msg_format("elu", strconv.Itoa(self.election.Elu)) + Msg_format("diffusion", "election")
			self.send_msg_targets(msg, self.voisins)
		} else if type_msg == "replicat" || type_msg == "tabTrainsToNewSite" {
			self.send_msg(self.election.Demandeur.ID().Address, rcvmsg) // Envoie du replicat ou du tableau de train au nouveau site
			mutex.Unlock()
			continue

		} else {
			self.initVague2(self.node.ID().Address)
			v := listVagues[self.node.ID().Address]
			v.Parent = self.node.ID().Address
			listVagues[self.node.ID().Address] = v
			msg = Msg_format("type", "bleu") + Msg_format("initiateurVague", self.node.ID().Address) + Msg_format("messageToCtrl", rcvmsg)
			self.send_msg_targets(msg, self.voisins)
		}

		mutex.Unlock()

		<-chNet // on traite le message prochain quand la vague termine

	}
}

func (self *App) receive_msg() {

	self.node.Handle(func(ctx noise.HandlerContext) error {
		mutex.Lock()
		data := string(ctx.Data())
		utils.Msg_receive(&self.p_nom, data)
		msg_type := Findval(data, "type")
		var min_fils = ""

		switch msg_type {
		case "copieTableauTermine":
			tableauCopieTermine--
			if tableauCopieTermine == 0 {
				tableauCopieTermine = 2
				utils.Msg_send(&self.p_nom, "Taille nbVoisins : "+strconv.Itoa(len(self.voisins)))
				if len(self.voisins) != 1 { // Les tableaux ont bien été copiés. Une autre election peut donc commencer.
					utils.Msg_send(&self.p_nom, "fin_admission")
					msg := Msg_format("type", "bleu") + Msg_format("admission", "fin_admission") + Msg_format("initiateurVague", self.node.ID().Address) + Msg_format("newSite", p_nom_demandeur)
					self.send_msg_targets(msg, filtrerVoisins(self.voisins, self.election.Demandeur.ID().Address))
				}
			}
		case "tabTrainsToNewSite":
			fmt.Println(data)                                                   // envoie du tableau de trains au controle
			self.send_msg(self.pere, Msg_format("type", "copieTableauTermine")) // dire au père que le nouveau a terminé de copier les tableaux
		case "replicat":
			fmt.Println(data)                                                   // envoie le replicat au controle
			self.send_msg(self.pere, Msg_format("type", "copieTableauTermine")) // dire au père que le nouveau a terminé de copier les tableaux
		case "changementParent":
			self.pere = Findval(data, "pere")
			self.voisins = filtrerVoisins(self.voisins, ctx.ID().Address)
			if self.pere != ":" {
				self.voisins = append(self.voisins, self.pere)
			}
			fmt.Println(Msg_format("type", "newNbVoisins") + Msg_format("nbVoisins", strconv.Itoa(len(self.voisins))))
			self.initElection()
			self.initVague()
		case "depart":
			self.recupererVoisins(Findval(data, "voisins"))
			self.voisins = filtrerVoisins(self.voisins, ctx.ID().Address)
			self.fils = filtrerVoisins(self.fils, ctx.ID().Address)
			if Findval(data, "racine") == ":" {
				self.pere = ":" // nouvelle racine
			}
			fmt.Println(Msg_format("type", "newNbVoisins") + Msg_format("nbVoisins", strconv.Itoa(len(self.voisins))))
			self.initElection()
			self.initVague()
			self.testVagueEnCours(ctx.ID().Address)
			v := listVagues[ctx.ID().Address]
			v.NbVoisinsAttendus++
			listVagues[ctx.ID().Address] = v // Le noeud qui part est encore dans le réseau et va diffuser "fin_admission"
		case "demande_admission":
			self.election.Demandeur = ctx
			p_nom_demandeur = Findval(data, "nom")
			if len(self.voisins) == 0 {
				self.accept_admission(self.election.Demandeur)
				self.initVague()
				self.initElection()
				fmt.Println(Msg_format("type", "nouveau_site") + Msg_format("new_numsite", p_nom_demandeur[1:]) + Msg_format("isFather", "true")) // Envoie au controle
				chNet <- 0
				//self.send_msg(self.election.Demandeur.ID().Address, rcvmsg) // Envoie du replicat au nouveau site
				break
			}
			if self.election.Parent == "" {
				self.election.Elu, _ = strconv.Atoi(self.node.Addr()[1:])
				self.election.Parent = self.node.Addr()[1:]
				utils.Msg_send(&self.p_nom, "demande_admission")
				msg := Msg_format("type", "bleu") + Msg_format("elu", strconv.Itoa(self.election.Elu)) + Msg_format("diffusion", "election")
				self.send_msg_targets(msg, self.voisins)
			}
		case "ok_admission":
			self.join_network(ctx.ID().Address)
			self.accepted_in_network = true
			self.initElection()
		case "bleu":
			if Findval(data, "diffusion") == "election" {
				k, _ := strconv.Atoi(Findval(data, "elu"))
				if self.election.Elu > k {
					demandeDepart = false
					self.election.Elu = k
					self.election.Parent = ctx.ID().Address
					self.election.NbVoisinsAtt--
					if self.election.NbVoisinsAtt > 0 {
						utils.Msg_send(&self.p_nom, "bleu_election")
						msg := Msg_format("type", "bleu") + Msg_format("elu", strconv.Itoa(self.election.Elu)) + Msg_format("diffusion", "election")
						self.send_msg_targets(msg, filtrerVoisins(self.voisins, self.election.Parent))
					} else {
						utils.Msg_send(&self.p_nom, "rouge_election")
						msg := Msg_format("type", "rouge") + Msg_format("elu", strconv.Itoa(self.election.Elu)) + Msg_format("diffusion", "election")
						self.send_msg(self.election.Parent, msg)
						self.initElection()
					}
				} else {
					if self.election.Elu == k {
						utils.Msg_send(&self.p_nom, "rouge_election")
						msg := Msg_format("type", "rouge") + Msg_format("elu", strconv.Itoa(self.election.Elu)) + Msg_format("diffusion", "election")
						self.send_msg(ctx.ID().Address, msg)
					}
				}
			} else {
				initiateurVague := Findval(data, "initiateurVague")
				self.testVagueEnCours(initiateurVague)
				vague := listVagues[initiateurVague]
				if vague.Parent == "" {
					admission := Findval(data, "admission")
					messageToCtrl := recoverMessageFromCtrl(data)
					numsiteDepart := Findval(data, "numsiteDepart")
					if admission == "fin_admission" {
						self.initElection()
						newSite := Findval(data, "newSite")
						if newSite != "" {
							fmt.Println(Msg_format("type", "nouveau_site") + Msg_format("new_numsite", newSite[1:]))
						} else if numsiteDepart != "" {
							fmt.Println(Msg_format("type", "departAutreSite") + Msg_format("numsiteDepart", numsiteDepart))
						}
					} else if messageToCtrl != "" {
						if utils.Findval(p_nom, data, "to") != "" {
							if string(utils.Findval(p_nom, data, "to")[1:]) == p_num {
								fmt.Println(messageToCtrl) // Envoie le message au controler si il y a "to" et qu'il m'est destiné
							}
						} else {
							fmt.Println(messageToCtrl) // Envoie le message au controler
						}
					}
					vague.Parent = ctx.ID().Address
					vague.NbVoisinsAttendus--
					listVagues[initiateurVague] = vague
					if vague.NbVoisinsAttendus > 0 {
						utils.Msg_send(&self.p_nom, "bleu_fin_admission")
						if admission == "fin_admission" {
							if numsiteDepart != "" {
								self.send_msg_targets(Msg_format("type", "bleu")+Msg_format("admission", "fin_admission")+Msg_format("initiateurVague", initiateurVague)+Msg_format("numsiteDepart", numsiteDepart), filtrerVoisins(self.voisins, vague.Parent))
							} else {
								self.send_msg_targets(Msg_format("type", "bleu")+Msg_format("admission", "fin_admission")+Msg_format("initiateurVague", initiateurVague)+Msg_format("newSite", Findval(data, "newSite")), filtrerVoisins(self.voisins, vague.Parent))
							}
						} else if messageToCtrl != "" {
							self.send_msg_targets(data, filtrerVoisins(self.voisins, vague.Parent))
						} else {
							self.send_msg_targets(Msg_format("type", "bleu")+Msg_format("initiateurVague", initiateurVague), filtrerVoisins(self.voisins, vague.Parent))
						}
					} else {
						utils.Msg_send(&self.p_nom, "rouge_fin_admission")
						self.send_msg(vague.Parent, Msg_format("type", "rouge")+Msg_format("initiateurVague", initiateurVague))
						self.initVague() //réinitialiser les variables pour une prochaine vague
						delete(listVagues, initiateurVague)
					}
				} else {
					utils.Msg_send(&self.p_nom, "rouge_vague")
					self.send_msg(vague.Parent, Msg_format("type", "rouge")+Msg_format("initiateurVague", initiateurVague))
				}
			}

		case "rouge":
			if Findval(data, "diffusion") == "election" {
				k, _ := strconv.Atoi(Findval(data, "elu"))
				if self.election.Elu == k {
					self.election.NbVoisinsAtt--
					if self.election.NbVoisinsAtt == 0 {
						i, _ := strconv.Atoi(self.node.Addr()[1:])
						if self.election.Elu == i {
							utils.Display_d(&self.p_nom, "rouge_election", "le site est élu")
							//fmt.Println("le site est élu")
							//FIN le site est élu
							parent = self.node.ID().Address
							if demandeDepart {
								if self.pere == ":" { // C'est la racine qui veut partir
									if len(self.fils) == 1 {
										utils.Msg_send(&self.p_nom, "transmission_voisins (CAS RACINE) nb_fils = 1")
										self.send_msg_targets(Msg_format("type", "changementParent")+Msg_format("pere", self.pere), self.fils)
										self.pere = self.fils[0]
									} else if len(self.fils) > 1 {
										utils.Msg_send(&self.p_nom, "transmission_voisins (CAS RACINE) nb_fils > 1")
										min_fils = minimum_fils(self.fils)
										self.send_msg(min_fils, Msg_format("type", "depart")+Msg_format("voisins", self.afficheVoisins())+Msg_format("racine", ":")) // pour dire à la nouvelle racine qu'elle devient racine
										self.send_msg_targets(Msg_format("type", "changementParent")+Msg_format("pere", min_fils), filtrerVoisins(self.fils, min_fils))
										self.pere = min_fils
									}
								} else {
									utils.Msg_send(&self.p_nom, "transmission_voisins")
									self.send_msg(self.pere, Msg_format("type", "depart")+Msg_format("voisins", self.afficheVoisins()))
									self.send_msg_targets(Msg_format("type", "changementParent")+Msg_format("pere", self.pere), self.fils)
								}
								time.Sleep(time.Duration(1) * time.Second)
							} else {
								self.accept_admission(self.election.Demandeur)
								fmt.Println(Msg_format("type", "nouveau_site") + Msg_format("new_numsite", p_nom_demandeur[1:]) + Msg_format("isFather", "true")) // Envoie au controle
								chNet <- 0
								// rcvmsg, _ := reader.ReadString('\n')                                                             // Le controle va envoyer son réplicat
								// rcvmsg = strings.TrimSpace(rcvmsg)
								// self.send_msg(self.election.Demandeur.ID().Address, rcvmsg) // Envoie du replicat au nouveau site
							}
							nbVoisinsAttendus-- // On ne doit pas compter le nouveau site arrivé
							//utils.Msg_send(&self.p_nom, "fin_admission")
							msg := Msg_format("type", "bleu") + Msg_format("admission", "fin_admission") + Msg_format("initiateurVague", self.node.ID().Address)
							if demandeDepart {
								str := *p_nom
								msg += Msg_format("numsiteDepart", str[1:])

								fmt.Println(Msg_format("type", "departAutreSite") + Msg_format("numsiteDepart", str[1:]))

								utils.Msg_send(&self.p_nom, "fin_admission")
								self.send_msg(self.pere, msg)
								utils.Msg_send(&self.p_nom, "controleur_quitter")
								fmt.Println(utils.Msg_format("from", self.name) + utils.Msg_format("type", "quit_ok_net"))
								time.Sleep(time.Duration(3) * time.Second)
								os.Exit(0)
							}
							self.initElection()
						} else {
							utils.Msg_send(&self.p_nom, "rouge_election")
							msg := Msg_format("type", "rouge") + Msg_format("elu", strconv.Itoa(self.election.Elu)) + Msg_format("diffusion", "election")
							self.send_msg(self.election.Parent, msg)
						}
					}
				}
			} else {
				initiateurVague := Findval(data, "initiateurVague")
				vague := listVagues[initiateurVague]
				vague.NbVoisinsAttendus--
				if vague.NbVoisinsAttendus == 0 {
					if vague.Parent != self.node.ID().Address {
						utils.Msg_send(&self.p_nom, "rouge_vague")
						self.send_msg(vague.Parent, Msg_format("type", "rouge")+Msg_format("initiateurVague", initiateurVague))
					}
					self.initVague() //réinitialiser les variables pour une prochaine vague
					delete(listVagues, initiateurVague)
					chNet <- 0 // Dire à receive_msg_from_ctrl que la vague est terminé il peut diffuser
					// d'autres messages
				}
			}

		default:
			utils.Display_w(&self.p_nom, "Receive_msg", "Type de message inconnu: "+data)
		}
		mutex.Unlock()
		return nil
	})
}

func minimum_fils(fils []string) string {
	min_fils := fils[0][1:]
	for i := 1; i < len(fils); i++ {
		if fils[i][1:] < min_fils {
			min_fils = fils[i][1:]
		}
	}
	min_fils = ":" + min_fils
	return min_fils
}

func (self *App) sendperiodic(cible string) {

	for {
		mutex.Lock()
		if self.accepted_in_network {
			mutex.Unlock()
			break
		}
		self.requete_admission(cible)
		mutex.Unlock()
		time.Sleep(time.Duration(5) * time.Second)
	}

	utils.Display_w(&self.p_nom, "sendperiodic", "Accepté dans le réseau !")

}

func (self *App) initVague() {
	parent = ""
	nbVoisinsAttendus = len(self.voisins)
}

func (self *App) initVague2(initiateurvague string) {
	v := listVagues[initiateurvague]
	v.NbVoisinsAttendus = len(self.voisins)
	v.Parent = ""
	listVagues[initiateurvague] = v
}

func (self *App) testVagueEnCours(initiateurvague string) {
	_, exists := listVagues[initiateurvague]
	if !exists {
		self.initVague2(initiateurvague)
	}
}

func (self *App) initElection() {
	self.election.Parent = ""
	self.election.NbVoisinsAtt = len(self.voisins)
	self.election.Elu = 99999999
}

func main() {

	flag.String("c", "", "cible")
	p_nom = flag.String("nom", "net", "nom")
	flag.Parse()
	cible := "localhost:" + flag.Lookup("c").Value.String()
	var err error

	s := *p_nom
	p_num = string(s[1])

	self := App{}
	self.name = *p_nom
	self.accepted_in_network = false
	self.voisins = []string{}
	self.node, err = noise.NewNode()
	self.p_nom = self.node.ID().Address
	self.pere = cible
	self.pere = strings.TrimPrefix(self.pere, "localhost")
	filtreVoisins := make([]string, 0)
	for _, voisin := range self.voisins {
		if voisin != self.pere {
			filtreVoisins = append(filtreVoisins, voisin)
		}
	}
	self.fils = filtreVoisins

	self.receive_msg()

	check(err)
	check(self.node.Listen())

	self.initVague()

	self.election = &Election{}
	self.initElection()

	if debug {
		fmt.Println("main: parent", parent, "nbVoisinsAttendus", nbVoisinsAttendus, "élu", self.election.Elu)
		fmt.Printf("Adresss %s\n", self.node.Addr()[1:])
	}

	if self.pere == "" {
		self.election.Elu, _ = strconv.Atoi(self.node.Addr()[1:])
	}

	utils.Display_w(&self.p_nom, "main", "Adresse actuelle: "+self.node.Addr())

	defer func(node *noise.Node) {
		err_close := node.Close()
		if err_close != nil {
			fmt.Println("Error while closing the node")
		}
	}(self.node)

	if cible != "localhost:" {
		self.sendperiodic(cible)
		//self.requete_admission(cible)
	}

	go self.receive_msg_from_ctrl()

	// maintient le programme actif
	select {}
}
