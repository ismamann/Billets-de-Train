#!/bin/bash

go build -o app app.go
go build -o controle controle.go
go build -o net net.go

num_app=${1}
cible=${2}


# Fonction de nettoyage
nettoyer() {
  echo "Nettoyage des FIFOs..."
  rm -f /tmp/in_A${num_app} /tmp/out_A${num_app}
  rm -f /tmp/in_C${num_app} /tmp/out_C${num_app}
  rm -f /tmp/in_N${num_app} /tmp/out_N${num_app}
  echo "FIFOs supprimés."
}

# # Si le script est appelé avec "nettoyer", exécute et quitte
if [ "$3" == "nettoyer" ]; then
  nettoyer
  exit 0
fi

# Exécute le nettoyage au début pour éviter des conflits de FIFOs existants
nettoyer

# Création des pipes nommés
mkfifo /tmp/in_A${num_app} /tmp/out_A${num_app}
mkfifo /tmp/in_C${num_app} /tmp/out_C${num_app}
mkfifo /tmp/in_N${num_app} /tmp/out_N${num_app}

# Taille plus grande : 100 colonnes x 30 lignes
SIZE="100x24"

# Lancement des applications dans des xterms
xterm -T "A${num_app}" -geometry ${SIZE}+120+290 -e "./app -nom A${num_app} -debug false" &

# Lancement des contrôleurs dans des xterms
xterm -T "C${num_app}" -geometry ${SIZE}+60+270 -e "./controle -nom C${num_app} -numero ${num_app} -nb_sites 1 -nb_trains 3 < /tmp/in_C${num_app} > /tmp/out_C${num_app}" &

# Lancement du net dans des xterms
xterm -T "N${num_app}" -geometry ${SIZE}+10+250 -e "./net -c ${cible} -nom N${num_app} < /tmp/in_N${num_app} > /tmp/out_N${num_app}" &

# Redirection : A{num_app} -> C{num_app}
(cat /tmp/out_A${num_app} > /tmp/in_C${num_app}) &

# Redirection : C{num_app} -> A{num_app} + N{num_app}
(tee /tmp/in_A${num_app} /tmp/in_N${num_app} < /tmp/out_C${num_app} > /dev/null) &

# Redirection : N{num_app} -> C{num_app}
(cat /tmp/out_N${num_app} > /tmp/in_C${num_app}) &

echo "Le nouveau processus est lancé dans un xterms."
