

if [ $# -ne 7 ]; then
  echo "Error: Please provide exactly 7 arguments for nodes"
  exit 1
fi


bin_name=$1
opengemini_log_path=$2
bin_file=$3
conf_file=$4
pid_file=$5
extra_file=$6
index=$7


pid=$(pgrep $bin_name)
[[ -n $pid ]] && kill -9 "$pid"
# if [[ -z $pid ]]; then
#   echo "no process needed to be killed"
# else
#   kill -9 "$pid"
# fi


sleep 3

# rm -rf $opengemini_log_path
mkdir -p $opengemini_log_path/$index

nohup $bin_file -config $conf_file -pidfile $pid_file> $extra_file 2>&1 &

echo "successfully start $bin_name"
