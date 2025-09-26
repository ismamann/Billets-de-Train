#!/bin/bash

go build -o app app.go
go build -o controle controle.go

# Fonction de nettoyage
nettoyer() {
  echo "Nettoyage des FIFOs..."
  rm -f /tmp/in_A1 /tmp/out_A1 /tmp/in_A2 /tmp/out_A2 /tmp/in_A3 /tmp/out_A3 /tmp/in_A4 /tmp/out_A4
  rm -f /tmp/in_C1 /tmp/out_C1 /tmp/in_C2 /tmp/out_C2 /tmp/in_C3 /tmp/out_C3 /tmp/in_C4 /tmp/out_C4
  echo "FIFOs supprimés."
}

# Si le script est appelé avec "nettoyer", exécute et quitte
if [ "$1" == "nettoyer" ]; then
  nettoyer
  exit 0
fi

# Exécute le nettoyage au début pour éviter des conflits de FIFOs existants
nettoyer

# Création des pipes nommés
mkfifo /tmp/in_A1 /tmp/out_A1
mkfifo /tmp/in_A2 /tmp/out_A2
mkfifo /tmp/in_C1 /tmp/out_C1
mkfifo /tmp/in_C2 /tmp/out_C2
mkfifo /tmp/in_A3 /tmp/out_A3
mkfifo /tmp/in_A4 /tmp/out_A4
mkfifo /tmp/in_C3 /tmp/out_C3
mkfifo /tmp/in_C4 /tmp/out_C4

# Lancement des applications dans des xterms
xterm -T "A1" -hold -e "./app -nom A1 -debug false" &
xterm -T "A2" -e "./app -nom A2 -debug false" &
xterm -T "A3" -e "./app -nom A3 -debug false" &
xterm -T "A4" -e "./app -nom A4 -debug false" &

# Lancement des contrôleurs dans des xterms
xterm -T "C1" -e "./controle -nom C1 -numero 1 -nb_sites 4 -nb_trains 3 < /tmp/in_C1 > /tmp/out_C1" &
xterm -T "C2" -e "./controle -nom C2 -numero 2 -nb_sites 4 -nb_trains 3 < /tmp/in_C2 > /tmp/out_C2" &
xterm -T "C3" -e "./controle -nom C3 -numero 3 -nb_sites 4 -nb_trains 3 < /tmp/in_C3 > /tmp/out_C3" &
xterm -T "C4" -e "./controle -nom C4 -numero 4 -nb_sites 4 -nb_trains 3 < /tmp/in_C4 > /tmp/out_C4" &

# Redirection : A1 -> C1
(cat /tmp/out_A1 > /tmp/in_C1) &

# Redirection : A2 -> C2
(cat /tmp/out_A2 > /tmp/in_C2) &

# Redirection : A3 -> C3
(cat /tmp/out_A3 > /tmp/in_C3) &

# Redirection : A4 -> C4
(cat /tmp/out_A4 > /tmp/in_C4) &

# Redirection : C1 -> A1 + C2 + C3 + C4
(tee /tmp/in_A1 /tmp/in_C2 /tmp/in_C3 /tmp/in_C4 < /tmp/out_C1 > /dev/null) &
#(tee /tmp/in_A1 /tmp/in_C2 < /tmp/out_C1 > /dev/null) &

# Redirection : C2 -> A2 + C1 + C4 + C3
(tee /tmp/in_A2 /tmp/in_C1 /tmp/in_C4 /tmp/in_C3 < /tmp/out_C2 > /dev/null) &
#(tee /tmp/in_A2 /tmp/in_C1 < /tmp/out_C2 > /dev/null) &

# Redirection : C3 -> A3 + C4 + C1 + C2
(tee /tmp/in_A3 /tmp/in_C4 /tmp/in_C1 /tmp/in_C2 < /tmp/out_C3 > /dev/null) &
#(tee /tmp/in_A3 /tmp/in_C4 < /tmp/out_C3 > /dev/null) &

# Redirection : C4 -> A4 + C3 + C2 + C1 
(tee /tmp/in_A4 /tmp/in_C3 /tmp/in_C2 /tmp/in_C1 < /tmp/out_C4 > /dev/null) &
#(tee /tmp/in_A4 /tmp/in_C3 < /tmp/out_C4 > /dev/null) &


echo "Tous les processus sont lancés dans des xterms."
