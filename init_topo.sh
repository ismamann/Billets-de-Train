#!/bin/bash

go build -o app app.go
go build -o controle controle.go
go build -o net net.go

# Nettoyage
nettoyer() {
  echo "Nettoyage des FIFOs..."
  rm -f /tmp/in_A1 /tmp/out_A1 
  rm -f /tmp/in_C1 /tmp/out_C1
  rm -f /tmp/in_N1 /tmp/out_N1
  echo "FIFOs supprimés."
}

if [ "$1" == "nettoyer" ]; then
  nettoyer
  exit 0
fi

nettoyer

# Création des FIFOs
mkfifo /tmp/in_A1 /tmp/out_A1
mkfifo /tmp/in_C1 /tmp/out_C1
mkfifo /tmp/in_N1 /tmp/out_N1

# Taille plus grande : 100 colonnes x 30 lignes
SIZE="100x24"

# Instance 1 - haut gauche (un peu descendue)
xterm -T "A1" -geometry ${SIZE}+120+290 -e "./app -nom A1 -debug false" &
xterm -T "C1" -geometry ${SIZE}+60+270 -e "./controle -nom C1 -numero 1 -nb_sites 1 -nb_trains 3 < /tmp/in_C1 > /tmp/out_C1" &
xterm -T "N1" -geometry ${SIZE}+10+250 -e "./net -nom N1 < /tmp/in_N1 > /tmp/out_N1" &

# Redirections app → controle
(cat /tmp/out_A1 > /tmp/in_C1) &

# Redirections controle → autres
(tee /tmp/in_A1 /tmp/in_N1 < /tmp/out_C1 > /dev/null) &

# Redirections net → controle
(cat /tmp/out_N1 > /tmp/in_C1) &

echo "Tous les processus sont lancés avec xterms."