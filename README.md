# README - Projet de Système Réparti pour la Simulation de la Vente de Billets de Train

Ce projet a pour objectif de concevoir un système distribué simulant la vente de billets de train, où différentes applications clientes et contrôleurs peuvent interagir et se synchroniser de manière décentralisée. Le système est structuré pour prendre en charge plusieurs trains, chacun étant traité comme une donnée répartie au sein du réseau.

## Cahier des charges

La première partie du projet portait sur la gestion des sites et des sauvegardes au sein d’un réseau fixe. La seconde partie visait à rendre ce réseau dynamique, en permettant l’intégration de nouveaux participants sans perturber son fonctionnement, tout en garantissant la cohérence globale.

Cela impliquait notamment la capacité à gérer l’arrivée et le départ des participants, à assurer la diffusion de messages avec terminaison explicite, ainsi qu’à résoudre les conflits éventuels à l’aide d’un mécanisme d’élection par extinction.


## Utilisation du Projet

#### Installatiion des dépendances
1. **Configuration:**
Il est nécessaire d'installer la dépendance avec `go get -u github.com/perlin-network/noise`. Il est impératif d'avoir le langage Go sur la machine ainsi que xterm. De plus le projet doit être executé sur un environnement UNIX.


2. **Execution:** 2 scripts shell permettent le bons fonctionnement du projet.
    -  init_topo.sh : Ce script permet de créé le premier site dans le réseau. C'est le premier script à lancer de cette façon :
    
    ```bash
    ./init_topo.sh
    ```
    
    - ajout_site_topo.sh : Celui-ci permet d'ajouter un site dans le réseau. Il prend 2 arguments, le numéro de l'instance et la cible à laquelle se raccrocher dans le réseau (une adresse). On l'utilise de cette façon :
    ```bash
    ./ajout_site_topo.sh 2 65412
    ```
    PS: on trouve le numéro de la "cible" sur le terminal du contrôleur net.


## Auteurs

Ce projet est développé par Dingwei Liu, Nuo Chen, Ismaël Driche et Alexandre Gauvin
