# bash filename value_for_node1 value_for_node2 value_for_node3

#!/usr/bin/env bash
#
#	Shell script to install openGemini as a cluster at one node.
#

if [ $# -ne 4 ]; then
  echo "Error: Please provide exactly 3 arguments for nodes"
  exit 1
fi

path=$1

echo $path

declare -a nodes[3]
nodes[1]=$2
nodes[2]=$3
nodes[3]=$4

if [ "$(uname)" == "Darwin" ]; then
  # generate config
  for((num = 1; num <= 3; num++))
  do
      rm -rf $path/etc/${nodes[num]}-openGemini.conf
      cp $path/v1.0.0/etc/openGemini.conf $path/etc/${nodes[num]}-openGemini.conf

      sed -i "" "s/{{meta_addr_1}}/${nodes[1]}/g" $path/etc/${nodes[num]}-openGemini.conf
      sed -i "" "s/{{meta_addr_2}}/${nodes[2]}/g" $path/etc/${nodes[num]}-openGemini.conf
      sed -i "" "s/{{meta_addr_3}}/${nodes[3]}/g" $path/etc/${nodes[num]}-openGemini.conf
      sed -i "" "s/{{addr}}/${nodes[$num]}/g" $path/etc/${nodes[num]}-openGemini.conf

      sed -i "" "s/{{id}}/$num/g" $path/etc/${nodes[num]}-openGemini.conf
  done
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
      # generate config
      for((num = 1; num <= 3; num++))
      do
          rm -rf $path/etc/${nodes[num]}-openGemini.conf
          cp $path/v1.0.0/etc/openGemini.conf $path/etc/${nodes[num]}-openGemini.conf

          sed -i "s/{{meta_addr_1}}/${nodes[1]}/g" $path/etc/${nodes[num]}-openGemini.conf
          sed -i "s/{{meta_addr_2}}/${nodes[2]}/g" $path/etc/${nodes[num]}-openGemini.conf
          sed -i "s/{{meta_addr_3}}/${nodes[3]}/g" $path/etc/${nodes[num]}-openGemini.conf
          sed -i "s/{{addr}}/${nodes[$num]}/g" $path/etc/${nodes[num]}-openGemini.conf

          sed -i "s/{{id}}/$num/g" $path/etc/${nodes[num]}-openGemini.conf
      done
else
  echo "not support the platform": $(uname)
  exit 1
fi