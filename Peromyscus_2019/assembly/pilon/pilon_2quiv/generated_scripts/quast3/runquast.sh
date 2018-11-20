#!/bin/bash
#$ -N quast3_pilon_1
#$ -pe openmp 32-64
#$ -R y
#$ -q bio,pub64,abio,free64,free48
#$ -ckpt restart
###$ -l kernel=blcr
###$ -r y
#$ -hold_jid pilon_1

cd $SGE_O_WORKDIR

#module load wgs
#module load MUMmer/3.23
#module load amos
#module load boost/1.49.0
#module load blat
#module load perl/5.16.2
#module load python/2.7.2


export PATH=/bio/jbaldwi1/dbg2olc/mel/pilon/test_pilon/correct_pilon/quast_run4_newquast/program/quast-3.2:$PATH

python /bio/jbaldwi1/dbg2olc/mel/pilon/test_pilon/correct_pilon/quast_run4_newquast/program/quast-3.2/quast.py ../pilon.fasta --debug -t ${CORES} -R /dfs1/bio/jbaldwi1/dbg2olc/mel/reference/dmel-all-chromosome-r6.01.fasta 1> quast_out.txt 2> quast_err.txt
