export PATH=/data1/jbrown/local_programs/anaconda/install_dir/anaconda/bin:$PATH
source activate py35

full_1dsq_basecaller.py -r -i /data1/jbrown/louse_project/raw_data/new_minion/1 -t 64 -s /data1/jbrown/louse_project/temp/minion_basecalling/1 -f FLO-MIN107 -k SQK-LSK308

source deactivate py35
