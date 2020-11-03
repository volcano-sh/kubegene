# GCS 流程设计

-----------------------------

## [Huawei Homepage](20200118_huawei.md)

## [Huawei Docker images](20200118_huawei_dockerimage.md)

## [Huawei 文件拷贝](20200118_huawei_copy.md)

## [Huawei GCS 流程设计](20200118_huawei_gcspipeline.md)

-----------------------------

## Content

* [进行数据拷贝](#进行数据拷贝) 
* [GCS 流程示意原理](#GCS流程示意原理)
* [两种流程执行的方式](#两种流程执行的方式)
* [文件挂载](#文件挂载)
* [流程描述](#流程描述)
* [RNA-seq toal](#RNA-seq-toal)
* [RNA-seq small](#RNA-seq-small)
* [RNA-seq ploy A](#RNA-seq-ploy-A)
* [Ribo-seq-Xtail](#Ribo-seq-Xtail) 
* [Ribo-seq Ribocode](#Ribo-seq-Ribocode) 
* [Ribo-seq RiboMiner](#Ribo-seq-RiboMiner) 
* [blastn](#blastn)  
* [blastp](#blastp)   
* [alternative splicing](#alternative-splicing)  
* [Chip-seq](#Chip-seq)  
* [bash-with-parameters](#bash-with-parameters)
* [对字符串进行分割成数组](#对字符串进行分割成数组) 

## 进行数据拷贝
```sh
obsutil config -i=5ULA********Y6P -k=gvroY******************02 -e=https://obs.cn-north-4.myhuaweicloud.com && obsutil cp -r -f -u obs://gene-container-xugang/gcs/huawei_file/a1-fastq/ /home/sfs && obsutil cp -r -f -u obs://gene-container-xugang/gcs/huawei_file/refrence/ /home/sfs && ls /home/sfs
```
https://support.huaweicloud.com/tr-gcs/gcs_tr_04_0002.html

## GCS流程示意原理
```sh
job-a:
  commands_iter:
    command: echo ${1} ${item}
    vars_iter:
      - [A, B, C]
```

```sh
如下示例中，commands有四行，则表示容器并发数量为4，每个容器分别执行不同的命令

commands:
  - sh /obs/gcscli/run-xxx/run.sh 1 a
  - sh /obs/gcscli/run-xxx/run.sh 2 a
  - sh /obs/gcscli/run-xxx/run.sh 1 b
  - sh /obs/gcscli/run-xxx/run.sh 2 b

  如果命令行是由多行组成，可以使用yaml语法中的“|”（保留换行符，整个字符串当做yaml中一个key的value）格式。这样就可以把大篇幅的命令行原封不动的拷贝过来，如：

commands:
  - |
    samtools merge -f -@ ${nthread} -b ${volume-path-sfs}/${sample}/mergelist.txt \
    ${volume-path-sfs}/${sample}/${sample}.sort.bam && \
    samtools flagstat ${volume-path-sfs}/${sample}/${sample}.sort.bam > ${volume-path-sfs}/${sample}/${sample}.sort.flagstat
```

## 两种流程执行的方式

看一下创建自定义流程的第5点
support.huaweicloud.com/bestpractice-gcs/gcs_bestpractice_001.html
看一下GCS_DATA_PVC这个变量的使用
support.huaweicloud.com/tr-gcs/gcs_tr_04_0004.html

```sh
   commands_iter:
      command: |
        bwa mem -t ${nthread} -M -R "@RG\tID:Sample\tPL:illumina\tSM:${sample}\tCN:GATK4" \
                ${volume-path-ref}/${reference-path}/${fastafile} \
                ${volume-path-sfs}/${sample}/${1}/R0.${1}.fastq  \
                ${volume-path-sfs}/${sample}/${1}/R1.${1}.fastq |\
        samtools view -F 4 -q 10 -bS /dev/stdin \
                >${volume-path-sfs}/${sample}/${sample}.${1}.bam && \
        samtools sort -@ ${nthread} \
                -o ${volume-path-sfs}/${sample}/${sample}.${1}.sort.bam \
                ${volume-path-sfs}/${sample}/${sample}.${1}.bam && \
        samtools index ${volume-path-sfs}/${sample}/${sample}.${1}.sort.bam
      vars_iter:
        - 'range(0, ${npart})'

    commands_iter:
      command: |
        if [ ! -f  ${volume-path-obs}/${1} ]; then echo "File ${volume-path-obs}/${1} not found" && exit 1; fi && \
        for i in `seq 0 ${npart}`; do mkdir -p ${volume-path-sfs}/${sample}/$i; done && \
        zlibfq -b 100000 -t ${npart} ${volume-path-obs}/${1} ${volume-path-sfs}/${sample} R${item}
      vars_iter:
        - '${fastq-files}'
```

## 流程描述
```sh
    description: '将数据拷贝到制定目录下/home/sfs'
    description: 将原始的下机数据去除接头
    description: 去除低质量的数据    
    description: 对数据进行质量评估
    description: 去除核糖体的reads
    description: 比对到基因组上
    description: 将结果拷贝回obsvolumn中

  (public-rnaseq-polya)
主要应用于分析 poly A 建库的下机数据。主要分为6个步骤，去除reads的接头，去除低质量的 reads, 生成reads 质量信息报告，去除核糖体序列，将reads 比对到基因组中，将文件移动归档。

inputs:
  JobName:
    type: string
    default: ployarnaseq
    description: 任务的名称
  AccessKey:
    type: string
    description: ak
    default: null
  SecretKey:
    type: string
    default: null
    description: sk
  endpoint:
    type: string
    default: 'https://obs.cn-north-4.myhuaweicloud.com'
    description: endpoint
  obs_location:
    type: string
    default: 'obs://hw-gcs-logo-cn-north-4-06a54be3938010610f01c00da675d700'
    description: obs桶的名称
  obs_data_path:
    type: string
    default: 'arabidopsis/huawei_file/a1-fastq/'
    description: 原始数据在桶中的相对路径。
  obs_reference_rRNA_bowtie:
    type: string
    default: 'arabidopsis/huawei_file/refrence/tair_rRNA_bowtie_index/tair.rRNA.fa'
    description: 物种的核糖体序列，bowtie建库数据的在obs中的相对路径
  obs_reference_genomeFile_star:
    type: string
    default: 'arabidopsis/huawei_file/refrence/tair_star/'
    description: 物种的基因组序列，STAR建库数据的在obs中的相对路径
  fastq_files:
    type: array
    default:
      - SRR3498212.fq
      - SRR966479.fq
    description: 原始数据文件名称
  adapter:
    type: string
    default: AAAAAAAAA
    description: read的接头序列
```

## RNA-seq toal
```sh

# cutadapter
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a2-cutadapter && \
        echo ${1} begin `date` /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        /home/obs/${obs_data_path}/${1} && \
        /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
         /home/obs/${obs_data_path}/${1}
      vars_iter:
        - '${fastq_files}'

# Fastx
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a3-filter && \
        echo ${1} begin `date` && \
        /home/test/bin/fastq_quality_filter \
        -Q33 -v -q 25 -p 75 \
        -i /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        -o /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq 
      vars_iter:
        - '${fastq_files}'

# fastQC
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a4-qc && \
        echo ${1} begin `date` && \
        /home/test/FastQC/fastqc \
        /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
        -o /home/sfs/${JobName}/a4-qc
      vars_iter:
        - '${fastq_files}'

# remove rRNA
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA && \
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA/nonrRNA && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
        /home/test/bowtie-1.2.3-linux-x86_64/bowtie \
        -n 0 -norc --best -l 15 -p 8 \
         --un=/home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} /home/obs/${obs_reference_rRNA_bowtie} \
         -q /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.alin > \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.err && \
         rm -rf /home/sfs/${JobName}/a5-rmrRNA/${1}.alin 
         
      vars_iter:
        - '${fastq_files}'

# mapping
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a6-map && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
         STAR --runThreadN 8 --alignEndsType EndToEnd \
         --outFilterMismatchNmax 1 --outFilterMultimapNmax 1 \
         --genomeDir /home/obs/${obs_reference_genomeFile_star} \
         --readFilesIn /home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} \
         --outFileNamePrefix /home/sfs/${JobName}/a6-map/${1} \
         --outSAMtype BAM SortedByCoordinate \
         --quantMode TranscriptomeSAM GeneCounts
         
      vars_iter:
        - '${fastq_files}'

# copy
  commands:
      - 'obsutil config -i=${AccessKey} -k=${SecretKey} -e=${endpoint}&& obsutil mkdir -p ${obs_location}/output/${JobName}/  && obsutil cp -r -f /home/sfs/${JobName}/ ${obs_location}/output/${JobName}/ && rm -rf /home/sfs/${JobName} && echo Check sfs && ls -al /home/sfs && ls -al /home/obs/output'
```

## RNA-seq small 
```sh
# cutadapter
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a2-cutadapter && \
        echo ${1} begin `date` /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        /home/obs/${obs_data_path}/${1} && \
        /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
         /home/obs/${obs_data_path}/${1}
      vars_iter:
        - '${fastq_files}'


# fastx
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a3-filter && \
        echo ${1} begin `date` && \
        /home/test/bin/fastq_quality_filter \
        -Q33 -v -q 25 -p 75 \
        -i /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        -o /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq 
      vars_iter:
        - '${fastq_files}'

# fastQC
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a4-qc && \
        echo ${1} begin `date` && \
        /home/test/FastQC/fastqc \
        /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
        -o /home/sfs/${JobName}/a4-qc
      vars_iter:
        - '${fastq_files}'

# remove rRNA
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA && \
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA/nonrRNA && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
        /home/test/bowtie-1.2.3-linux-x86_64/bowtie \
        -n 0 -norc --best -l 15 -p 8 \
         --un=/home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} /home/obs/${obs_reference_rRNA_bowtie} \
         -q /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.alin > \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.err && \
         rm -rf /home/sfs/${JobName}/a5-rmrRNA/${1}.alin 
         
      vars_iter:
        - '${fastq_files}'

# bowtie mature
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a7-mature && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
        /home/test/bowtie-1.2.3-linux-x86_64/bowtie \
        -n 0 -norc --best -l 15 -p 8 \
        /home/obs/${obs_reference_mature_bowtie} \
         -q /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
         /home/sfs/${JobName}/a7-mature/${1}.alin > \
         /home/sfs/${JobName}/a7-mature/${1}.err

         
      vars_iter:
        - '${fastq_files}'

# bowtie hairpin
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a8-hairpin && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
        /home/test/bowtie-1.2.3-linux-x86_64/bowtie \
        -n 0 -norc --best -l 15 -p 8 \
         /home/obs/${obs_reference_hairpin_bowtie} \
         -q /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
         /home/sfs/${JobName}/a8-hairpin/${1}.alin > \
         /home/sfs/${JobName}/a8-hairpin/${1}.err
         
      vars_iter:
        - '${fastq_files}'

#  STAR
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a6-map && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
         STAR --runThreadN 8 --alignEndsType EndToEnd \
         --outFilterMismatchNmax 1 --outFilterMultimapNmax 1 \
         --genomeDir /home/obs/${obs_reference_genomeFile_star} \
         --readFilesIn /home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} \
         --outFileNamePrefix /home/sfs/${JobName}/a6-map/${1} \
         --outSAMtype BAM SortedByCoordinate \
         --quantMode TranscriptomeSAM GeneCounts
         
      vars_iter:
        - '${fastq_files}'

# copy
    commands:
      - 'obsutil config -i=${AccessKey} -k=${SecretKey} -e=${endpoint}&& obsutil mkdir -p ${obs_location}/output/${JobName}/  && obsutil cp -r -f /home/sfs/${JobName}/ ${obs_location}/output/${JobName}/ && rm -rf /home/sfs/${JobName} && echo Check sfs && ls -al /home/sfs && ls -al /home/obs/output'

```

## RNA-seq ploy A
```sh
# cutadapter
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a2-cutadapter && \
        echo ${1} begin `date` /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        /home/obs/${obs_data_path}/${1} && \
        /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
         /home/obs/${obs_data_path}/${1}
      vars_iter:
        - '${fastq_files}'

# fastx
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a3-filter && \
        echo ${1} begin `date` && \
        /home/test/bin/fastq_quality_filter \
        -Q33 -v -q 25 -p 75 \
        -i /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        -o /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq 
      vars_iter:
        - '${fastq_files}'

# fastqc
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a4-qc && \
        echo ${1} begin `date` && \
        /home/test/FastQC/fastqc \
        /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
        -o /home/sfs/${JobName}/a4-qc
      vars_iter:
        - '${fastq_files}'

# remove rRNA
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA && \
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA/nonrRNA && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
        /home/test/bowtie-1.2.3-linux-x86_64/bowtie \
        -n 0 -norc --best -l 15 -p 8 \
         --un=/home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} /home/obs/${obs_reference_rRNA_bowtie} \
         -q /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.alin > \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.err && \
         rm -rf /home/sfs/${JobName}/a5-rmrRNA/${1}.alin 
         
      vars_iter:
        - '${fastq_files}'

# STAR
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a6-map && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
         STAR --runThreadN 8 --alignEndsType EndToEnd \
         --outFilterMismatchNmax 1 --outFilterMultimapNmax 1 \
         --genomeDir /home/obs/${obs_reference_genomeFile_star} \
         --readFilesIn /home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} \
         --outFileNamePrefix /home/sfs/${JobName}/a6-map/${1} \
         --outSAMtype BAM SortedByCoordinate \
         --quantMode TranscriptomeSAM GeneCounts
         
      vars_iter:
        - '${fastq_files}'

# copy
    commands:
      - 'obsutil config -i=${AccessKey} -k=${SecretKey} -e=${endpoint}&& obsutil mkdir -p ${obs_location}/output/${JobName}/  && obsutil cp -r -f /home/sfs/${JobName}/ ${obs_location}/output/${JobName}/ && rm -rf /home/sfs/${JobName} && echo Check sfs && ls -al /home/sfs && ls -al /home/obs/output'

```

## Ribo-seq Xtail

```sh
#cutadapter
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a2-cutadapter && \
        ls /home/obs/${obs_data_path}/ && \
        echo ${1} begin `date` /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        /home/sfs/${JobName}/a1-fastq/${1} && \
        /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
         /home/obs/${obs_data_path}/${1}
    vars_iter:
        - '${fastq_files}'

# fastq_quality_filter
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a3-filter && \
        echo ${1} begin `date` && \
        /home/test/bin/fastq_quality_filter \
        -Q33 -v -q 25 -p 75 \
        -i /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        -o /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq 
      vars_iter:
        - '${fastq_files}'

# fastQC
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a4-qc && \
        echo ${1} begin `date` && \
        /home/test/FastQC/fastqc \
        /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
        -o /home/sfs/${JobName}/a4-qc
      vars_iter:
        - '${fastq_files}'

#remove rRNA
            commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA && \
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA/nonrRNA && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
        /home/test/bowtie-1.2.3-linux-x86_64/bowtie \
        -n 0 -norc --best -l 15 -p 8 \
         --un=/home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} /home/obs/${obs_reference_rRNA_bowtie} \
         -q /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.alin > \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.err && \
         rm -rf /home/sfs/${JobName}/a5-rmrRNA/${1}.alin 
         
      vars_iter:
        - '${fastq_files}'

#STAR
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a6-map && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
         STAR --runThreadN 8 --alignEndsType EndToEnd \
         --outFilterMismatchNmax 1 --outFilterMultimapNmax 1 \
         --genomeDir /home/obs/${obs_reference_genomeFile_star} \
         --readFilesIn /home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} \
         --outFileNamePrefix /home/sfs/${JobName}/a6-map/${1} \
         --outSAMtype BAM SortedByCoordinate \
         --quantMode TranscriptomeSAM GeneCounts
         
      vars_iter:
        - '${fastq_files}'

# htseq
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/ && \
        mkdir -p /home/sfs/${JobName}/a7-htcount && \
        ls /root/miniconda2/bin/ && \
        ls /home/sfs/${JobName}/a6-map/ &&\
        echo ${1} begin `date` && \
        echo /root/miniconda2/bin/python \
        /root/miniconda2/bin/RPF_count_CDS.py \
        /home/sfs/${JobName}/a6-map/${1}Aligned.sortedByCoord.out.bam \
        /home/obs/${obs_reference_gtf} > /home/sfs/${JobName}/a7-htcount/${1}.count && \
        /root/miniconda2/bin/python \
        /root/miniconda2/bin/RPF_count_CDS.py \
        /home/sfs/${JobName}/a6-map/${1}Aligned.sortedByCoord.out.bam \
        /home/obs/${obs_reference_gtf} > /home/sfs/${JobName}/a7-htcount/${1}.count
      vars_iter:
        - '${fastq_files}'

# merge
commands:
      - ' ls -alh /home/sfs/${JobName}/a7-htcount && echo bash  /root/miniconda2/bin/merge.sh -a ${fastq_files_name} -b ${fastq_files_label} -c /home/sfs/${JobName}/a7-htcount  &&  bash  /root/miniconda2/bin/merge.sh -a ${fastq_files_name} -b ${fastq_files_label} -c /home/sfs/${JobName}/a7-htcount '

# xtail
        commands:
      - ' mkdir -p /home/sfs/${JobName}/a8-xtail && Rscript /home/test/xtail.r /home/sfs/${JobName}/a7-htcount/merge.counter ${xtail_ribo_vector} ${xtail_rna_vector} ${xtail_label} /home/sfs/${JobName}/a8-xtail '

# cp
commands:
      - 'obsutil config -i=${AccessKey} -k=${SecretKey} -e=${endpoint}&& obsutil mkdir -p ${obs_location}/output/${JobName}/  && obsutil cp -r -f /home/sfs/${JobName}/ ${obs_location}/output/${JobName}/ && rm -rf /home/sfs/${JobName} && echo Check sfs && ls -al /home/sfs && ls -al /home/obs/output'
   
```

## Ribo-seq Ribocode

```sh
#cutadapter
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a2-cutadapter && \
        ls /home/obs/${obs_data_path}/ && \
        echo ${1} begin `date` /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        /home/sfs/${JobName}/a1-fastq/${1} && \
        /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
         /home/obs/${obs_data_path}/${1}
    vars_iter:
        - '${fastq_files}'

# fastq_quality_filter
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a3-filter && \
        echo ${1} begin `date` && \
        /home/test/bin/fastq_quality_filter \
        -Q33 -v -q 25 -p 75 \
        -i /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        -o /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq 
      vars_iter:
        - '${fastq_files}'

# fastQC
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a4-qc && \
        echo ${1} begin `date` && \
        /home/test/FastQC/fastqc \
        /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
        -o /home/sfs/${JobName}/a4-qc
      vars_iter:
        - '${fastq_files}'

#remove rRNA
            commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA && \
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA/nonrRNA && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
        /home/test/bowtie-1.2.3-linux-x86_64/bowtie \
        -n 0 -norc --best -l 15 -p 8 \
         --un=/home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} /home/obs/${obs_reference_rRNA_bowtie} \
         -q /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.alin > \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.err && \
         rm -rf /home/sfs/${JobName}/a5-rmrRNA/${1}.alin 
         
      vars_iter:
        - '${fastq_files}'

#STAR
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a6-map && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
         STAR --runThreadN 8 --alignEndsType EndToEnd \
         --outFilterMismatchNmax 1 --outFilterMultimapNmax 1 \
         --genomeDir /home/obs/${obs_reference_genomeFile_star} \
         --readFilesIn /home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} \
         --outFileNamePrefix /home/sfs/${JobName}/a6-map/${1} \
         --outSAMtype BAM SortedByCoordinate \
         --quantMode TranscriptomeSAM GeneCounts \
         --limitBAMsortRAM 7110037687 
         
      vars_iter:
        - '${fastq_files}'

# Ribocode
    commands:
      - 'mkdir -p /home/sfs/${JobName}/a7-ribocode_annotation && /root/miniconda3/bin/prepare_transcripts -g /home/obs/${obs_reference_gtf} -f  /home/obs/${obs_reference_fasta} -o /home/sfs/${JobName}/a7-ribocode_annotation && mkdir -p /home/sfs/${JobName}/a8-ribocode  '

    commands_iter:
      command: |
        head /home/obs/output/riboseq-ribocode/riboseq-ribocode/a6-map/${1}Aligned.toTranscriptome.out.bam && echo "/home/obs/output/riboseq-ribocode/riboseq-ribocode/a6-map/${1}Aligned.toTranscriptome.out.bam" > /home/sfs/${JobName}/a8-ribocode/ltrans.${1}.txt;
      vars_iter:
        - '${fastq_files}'

    commands:
      - >-
        cat /home/sfs/${JobName}/a8-ribocode/ltrans.*.txt > /home/sfs/${JobName}/a8-ribocode/transcript.txt && rm -rf /home/sfs/${JobName}/a8-ribocode/ltrans.*.txt && cat /home/sfs/${JobName}/a8-ribocode/transcript.txt && /root/miniconda3/bin/metaplots -a /home/sfs/${JobName}/a7-ribocode_annotation
        -i /home/sfs/${JobName}/a8-ribocode/transcript.txt -o /home/sfs/${JobName}/a8-ribocode/a && mkdir -p /home/sfs/${JobName}/a9-ribocode-result && /root/miniconda3/bin/RiboCode -a /home/sfs/${JobName}/a7-ribocode_annotation -c /home/sfs/${JobName}/a8-ribocode/a_pre_config.txt -l no -g -o
        /home/sfs/${JobName}/a9-ribocode-result/        

# cp
commands:
      - 'obsutil config -i=${AccessKey} -k=${SecretKey} -e=${endpoint}&& obsutil mkdir -p ${obs_location}/output/${JobName}/  && obsutil cp -r -f /home/sfs/${JobName}/ ${obs_location}/output/${JobName}/ && rm -rf /home/sfs/${JobName} && echo Check sfs && ls -al /home/sfs && ls -al /home/obs/output'
   
```

### test on hub docker.
 
```sh
docker run --name=ribocode -dt -v /lulab/lustre2/xugang/docker_backup:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/ribocode_ribominer:1.0
docker exec -it ribocode bash
#
# make annoatation
mkdir -p /home/sfs/a7-ribocode_annotation2 && /root/miniconda3/bin/prepare_transcripts -g /home/sfs/huawei/Arabidopsis_thaliana.TAIR10.43.gtf -f /home/sfs/huawei/Arabidopsis_thaliana.TAIR10.dna.toplevel.fa.clean.fa -o /home/sfs/a7-ribocode_annotation2
# generate bam file
mkdir -p /home/sfs/a8-ribocode &&  rm -rf /home/sfs/a8-ribocode/transcript.txt &&  IFS=';' read -ra name <<< "SRR1958702;SRR1958703;SRR1958704;SRR3498206" && for i in ${name[@]}; do echo "/home/sfs/huawei/"$i".Aligned.toTranscriptome.out.bam" >> /home/sfs/a8-ribocode/transcript.txt;done;
# set p site
/root/miniconda3/bin/metaplots -a /home/sfs/a7-ribocode_annotation -i /home/sfs/a8-ribocode/transcript.txt -o /home/sfs/a8-ribocode/a
#ribocode
mkdir -p /home/sfs/a9-ribocode-result && /root/miniconda3/bin/RiboCode -a /home/sfs/a7-ribocode_annotation -c /home/sfs/a8-ribocode/a_pre_config.txt -l no -g -o /home/sfs/a9-ribocode-result/

exit

```

## 文件挂载

```sh
volumes:
  volumes-4ndk:
    mount_path: '/home/sfs'
    mount_from:
      pvc: '${GCS_SFS_PVC}'
  genobs:
    mount_path: '/home/obs'
    mount_from:
      pvc: '${GCS_DATA_PVC}'
```

## Ribo-seq RiboMiner

**Quality Control (QC):** Quality control for ribosome profiling data, containing periodicity checking, reads distribution among different reading frames,length distribution of ribosome footprints and DNA contaminations.
**Metagene Analysis (MA):** Metagene analysis among different samples to find possible ribosome stalling events.
**Feature Analysis (FA):** Feature analysis among different gene sets identified in MA step to explain the possible ribosome stalling.
**Enrichment Analysis (EA):** Enrichment analysis to find possible co-translation events.

### version 20200701
### fajin数据 Quality Control (QC)，Metagene Analysis (MA) 完成。
```sh
# annotation, prepare transcript;output transcript info; get protein coding sequence; Get URT sequence.
    commands:
      - >-
        mkdir -p /home/sfs/${JobName}/a7-RiboCode_annot &&  /root/miniconda3/bin/prepare_transcripts -g /home/obs/${obs_reference_gtf} -f /home/obs/${obs_reference_fasta} -o /home/sfs/${JobName}/a7-RiboCode_annot && mkdir -p /home/sfs/${JobName}/a8-Ribominer_annot && 
        /root/miniconda3/bin/OutputTranscriptInfo -c /home/sfs/${JobName}/a7-RiboCode_annot/transcripts_cds.txt -g /home/obs/${obs_reference_gtf} -f /home/sfs/${JobName}/a7-RiboCode_annot/transcripts_sequence.fa -o /home/sfs/${JobName}/a8-Ribominer_annot/longest.transcripts.info.txt -O
        /home/sfs/${JobName}/a8-Ribominer_annot/all.transcripts.info.txt && /root/miniconda3/bin/GetProteinCodingSequence -i /home/sfs/${JobName}/a7-RiboCode_annot/transcripts_sequence.fa  -c /home/sfs/${JobName}/a8-Ribominer_annot/longest.transcripts.info.txt -o
        /home/sfs/${JobName}/a8-Ribominer_annot/transcript --mode whole --table 1 && /root/miniconda3/bin/GetUTRSequences -i /home/sfs/${JobName}/a8-Ribominer_annot/transcript_transcript_sequences.fa -o /home/sfs/${JobName}/a8-Ribominer_annot/utr -c
        /home/sfs/${JobName}/a7-RiboCode_annot/transcripts_cds.txt
# metaplots.
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a9-metaplots && /root/miniconda3/bin/metaplots -a /home/sfs/${JobName}/a7-RiboCode_annot -r ${bam_files}/${1}Aligned.toTranscriptome.out.bam -o /home/sfs/${JobName}/a9-metaplots/${1}
      vars_iter:
        - '${fastq_files}'
# bam sort, index.
    commands_iter:
      command: >
        samtools sort -T ${bam_files}/${1}Aligned.toTranscriptome.tmp.bam -o ${bam_files}/${1}Aligned.toTranscriptome.out.sorted.bam ${bam_files}/${1}Aligned.toTranscriptome.out.bam && samtools index ${bam_files}/${1}Aligned.toTranscriptome.out.sorted.bam && samtools index
        ${bam_files}/${1}Aligned.sortedByCoord.out.bam #
      vars_iter:
        - '${fastq_files}'

# perodicity
    commands_iter:
      command: >
        mkdir -p /home/sfs/${JobName}/a10-periodicity && /root/miniconda3/bin/Periodicity -i ${bam_files}/${1}Aligned.toTranscriptome.out.sorted.bam -a /home/sfs/${JobName}/a7-RiboCode_annot -o /home/sfs/${JobName}/a10-periodicity/${1}_periodicity -c
        /home/sfs/${JobName}/a8-Ribominer_annot/longest.transcripts.info.txt -L 25 -R 35
      vars_iter:
        - '${fastq_files}'
# generate attributes.txt
    commands:
      - >-
        cd ${home_dir}/a9-metaplots/ && echo -e "#SampleName\tAlignmentFile\tStranded\tReadLength\tP-site" > attributes.txt && for i in `ls |grep _pre_config.txt`;do echo $i;grep -v "#" ${i}|grep .>> attributes.txt ;done && awk 'BEGIN {FS="\t"; OFS="\t"} {print $2, $4, $5, $1}' attributes.txt > 
        attributes2.txt && mv  attributes2.txt  attributes.txt && sed -i 's/Aligned.toTranscriptome.out.bam/Aligned.toTranscriptome.out.sorted.bam/g' attributes.txt && cut -f 1 ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt |sed '1d'  > ${home_dir}/a8-Ribominer_annot/select_trans.txt ;

# RiboDensityOfDiffFrames.
    commands:
      - 'mkdir -p ${home_dir}/a11-ribodensity && /root/miniconda3/bin/RiboDensityOfDiffFrames -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a11-ribodensity/a6-ribo-density-diff-frame'
    
# StatisticReadsOnDNAcontam
    commands_iter:
      command: |
        mkdir -p ${home_dir}/a12-dna-contamination && /root/miniconda3/bin/StatisticReadsOnDNAsContam -i  ${bam_files}/${1}Aligned.sortedByCoord.out.bam  -g /home/obs/${obs_reference_gtf} -o  ${home_dir}/a12-dna-contamination/${1}
      vars_iter:
        - '${fastq_files}'

######################
# Metagene Analysis
######################  
# MetageneAnalysisForTheWholeRegions; PlotMetageneAnalysisForTheWholeRegions
    commands:
      - >-
        mkdir -p  ${home_dir}/a13-metagene && /root/miniconda3/bin/MetageneAnalysisForTheWholeRegions -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a13-metagene/a8-metagene -b 15,90,60 -l 100 -n 10 -m 1 -e 5 --plot yes &&
        /root/miniconda3/bin/PlotMetageneAnalysisForTheWholeRegions -i ${home_dir}/a13-metagene/a8-metagene_scaled_density_dataframe.txt -o ${home_dir}/a13-metagene/a9-meta_gene_whole_regin -g ${gname} -r ${rname} -b 15,90,60 --mode all    
# MetageneAnalysis
    commands:
      - >-
        mkdir -p ${home_dir}/a14-metageneAnalysis && /root/miniconda3/bin/MetageneAnalysis -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a14-metageneAnalysis/b1-meat-cds -U codon -M RPKM -u 0 -d 500 -l 100 -n 10 -m 1 -e 5
        --norm yes -y 100 --CI 0.95 --type CDS && /root/miniconda3/bin/MetageneAnalysis -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a14-metageneAnalysis/b2-meat-utr -U nt -M RPKM -u 100 -d 100 -l 100 -n 10 -m 1 -e 5 --norm yes -y 50 --CI 0.95 --type UTR
# PolarityCalculation 
    commands:
      - >-
        mkdir -p ${home_dir}/a15-polarity && /root/miniconda3/bin/PolarityCalculation -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a15-polarity/b3-polarity -n 64 && /root/miniconda3/bin/PlotPolarity -i
        ${home_dir}/a15-polarity/b3-polarity_polarity_dataframe.txt -o ${home_dir}/a15-polarity/b4-plotpolarity -g ${gname} -r ${rname} -y 5 

##########################
# Feature Analysis (FA)
##########################        
# RiboDensityForSpecificRegion; RiboDensityAtEachKindAAOrCodon; PlotRiboDensityAtEachKindAAOrCodon
    commands:
      - >-
        mkdir -p ${home_dir}/a16-ribodensitycodon &&  /root/miniconda3/bin/RiboDensityForSpecificRegion -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a16-ribodensitycodon/b5-transcript-enrich -U codon -M RPKM -L 25 -R 75  &&
        /root/miniconda3/bin/RiboDensityAtEachKindAAOrCodon -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a16-ribodensitycodon/b6-ribosome-aa -M counts -S ${home_dir}/a8-Ribominer_annot/select_trans.txt  -l 100 -n 10 --table
        1 -F ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa &&  /root/miniconda3/bin/PlotRiboDensityAtEachKindAAOrCodon -i ${home_dir}/a16-ribodensitycodon/b6-ribosome-aa_all_codon_density.txt -o ${home_dir}/a16-ribodensitycodon/b7-PlotRiboDensityAtEachKindAAOrCodon -g ${gname} -r
        ${rname} --level AA

# PausingScore
    commands:
      - >-
        mkdir -p ${home_dir}/a17-PausingScore && /root/miniconda3/bin/PausingScore -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o  ${home_dir}/a17-PausingScore/b8-PausingScore -M counts -S ${home_dir}/a8-Ribominer_annot/select_trans.txt -l 100 -n 10 --table 1 -F  ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa && /root/miniconda3/bin/ProcessPausingScore -i ${pause_name} -o ${home_dir}/a17-PausingScore/b9-ProcessPausingScore -g ${gname} -r ${rname} --mode raw --ratio_filter 0 --pausing_score_filter 0
# move data.
    commands:
      - 'mv /home/sfs/${JobName} /home/obs/output/${JobName}/ '

```
ribominer-Metagene-Analysis.sh(fajin 0708) 
```sh
######################
# Metagene Analysis
######################  
# MetageneAnalysisForTheWholeRegions; PlotMetageneAnalysisForTheWholeRegions
    commands:
      - >-
        mkdir -p  ${home_dir}/a13-metagene && cd /home/obs/fjdata/data && /root/miniconda3/bin/MetageneAnalysisForTheWholeRegions -f /home/obs/${attributes} -c /home/obs/${longest_tra} -o ${home_dir}/a13-metagene/a8-metagene -b 15,90,60 -l 100 -n 10 -m 1 -e 5 --plot yes &&
        /root/miniconda3/bin/PlotMetageneAnalysisForTheWholeRegions -i ${home_dir}/a13-metagene/a8-metagene_scaled_density_dataframe.txt -o ${home_dir}/a13-metagene/a9-meta_gene_whole_regin -g ${gname} -r ${rname} -b 15,90,60 --mode all    
# MetageneAnalysis
    commands:
      - >-
        mkdir -p ${home_dir}/a14-metageneAnalysis && cd /home/obs/fjdata/data && /root/miniconda3/bin/MetageneAnalysis -f /home/obs/${attributes} -c /home/obs/${longest_tra} -o ${home_dir}/a14-metageneAnalysis/b1-meat-cds -U codon -M RPKM -u 0 -d 500 -l 100 -n 10 -m 1 -e 5
        --norm yes -y 100 --CI 0.95 --type CDS && /root/miniconda3/bin/MetageneAnalysis -f /home/obs/${attributes} -c /home/obs/${longest_tra} -o ${home_dir}/a14-metageneAnalysis/b2-meat-utr -U nt -M RPKM -u 100 -d 100 -l 100 -n 10 -m 1 -e 5 --norm yes -y 50 --CI 0.95 --type UTR
# PolarityCalculation 
    commands:
      - >-
        mkdir -p ${home_dir}/a15-polarity && cd /home/obs/fjdata/data && /root/miniconda3/bin/PolarityCalculation -f /home/obs/${attributes} -c /home/obs/${longest_tra} -o ${home_dir}/a15-polarity/b3-polarity -n 64 && /root/miniconda3/bin/PlotPolarity -i
        ${home_dir}/a15-polarity/b3-polarity_polarity_dataframe.txt -o ${home_dir}/a15-polarity/b4-plotpolarity -g ${gname} -r ${rname} -y 5 
```

ribominer-feature-analysis
```sh
##########################
# Feature Analysis (FA)
##########################        
# RiboDensityForSpecificRegion; RiboDensityAtEachKindAAOrCodon; PlotRiboDensityAtEachKindAAOrCodon
    commands:
      - >-
        mkdir -p /home/obs/${obs_dir}/a16-ribodensitycodon2 &&  /root/miniconda3/bin/RiboDensityForSpecificRegion -f /home/obs/${obs_dir}/a9-metaplots/attributes.txt -c /home/obs/${longest_tra} -o /home/obs/${obs_dir}/a16-ribodensitycodon2/b5-transcript-enrich -U codon -M RPKM -L 25 -R 75  &&
        /root/miniconda3/bin/RiboDensityAtEachKindAAOrCodon -f /home/obs/${obs_dir}/a9-metaplots/attributes.txt -c /home/obs/${longest_tra} -o /home/obs/${obs_dir}/a16-ribodensitycodon2/b6-ribosome-aa -M counts  -l 100 -n 10 --table
        1 -F /home/obs/${trans_cds_seq} &&  /root/miniconda3/bin/PlotRiboDensityAtEachKindAAOrCodon -i /home/obs/${obs_dir}/a16-ribodensitycodon2/b6-ribosome-aa_all_codon_density.txt -o /home/obs/${obs_dir}/a16-ribodensitycodon2/b7-PlotRiboDensityAtEachKindAAOrCodon -g ${gname} -r
        ${rname} --level AA

# Ribosome density around the triplete amino acid (tri-AA) motifs.
#
    commands:
      - >-
      mkdir -p /home/obs/${obs_dir}/a18-AroundTriplete/ && 
      /root/miniconda3/bin/RiboDensityAroundTripleteAAMotifs -f /home/obs/${obs_dir}/a9-metaplots/attributes.txt -c /home/obs/${longest_tra} -o /home/obs/${obs_dir}/a18-AroundTriplete/c0-RiboDensityAroundTripleteAAMotifs_PPP -M counts -l 100 -n 10 --table 1 -F /home/obs/${trans_cds_seq} --type2 PPP --type1 PP && 
      /root/miniconda3/bin/PlotRiboDensityAroundTriAAMotifs -i /home/obs/${obs_dir}/a18-AroundTriplete/c0-RiboDensityAroundTripleteAAMotifs_PPP_motifDensity_dataframe.txt -o /home/obs/${obs_dir}/a18-AroundTriplete/c1-PPP_plot -g ${gname} -r ${rname} --mode mean
# PausingScore
    commands:
      - >-
        mkdir -p /home/obs/${obs_dir}/a17-PausingScore2 && /root/miniconda3/bin/PausingScore -f /home/obs/${obs_dir}/a9-metaplots/attributes.txt -c /home/obs/${longest_tra} -o  /home/obs/${obs_dir}/a17-PausingScore2/b8-PausingScore -M counts -l 100 -n 10 --table 1 -F  /home/obs/${trans_cds_seq} && /root/miniconda3/bin/ProcessPausingScore -i ${pause_name} -o /home/obs/${obs_dir}/a17-PausingScore2/b9-ProcessPausingScore -g ${gname} -r ${rname} --mode raw --ratio_filter 0 --pausing_score_filter 0
# RPFdist calculation. GCContent
          commands:
      - >-
      mkdir -p ${home_dir}/a20-RPFdist-GCContent/ && 
          /root/miniconda3/bin/RPFdist -f ${home_dir}/a9-metaplots/attributes.txt -c /home/obs/${longest_tra} -o ${home_dir}/a20-RPFdist-GCContent/c3-RPFdist -M counts -l 100 -n 10 -m 1 -e 5 &&
          /root/miniconda3/bin/GCContent -i /home/obs/${trans_cds_seq} -o ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-normal --mode normal &&
          /root/miniconda3/bin/GCContent -i /home/obs/${trans_cds_seq} -o ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-frames --mode frames &&
          /root/miniconda3/bin/PlotGCContent -i ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-normal_GC_content.txt -o ${home_dir}/a20-RPFdist-GCContent/c5-PlotGCContent-normal --mode normal &&
          /root/miniconda3/bin/PlotGCContent -i ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-frames_GC_content_frames.txt -o ${home_dir}/a20-RPFdist-GCContent/c5-PlotGCContent-frames --mode frames

# Local tRNA adaptation index and global tRNA adaptation index
    commands:
      - >-  
        mkdir -p ${home_dir}/a21-tAI-cAI/ && cd /home/obs/${transcript_location} && /root/miniconda3/bin/tAI -i up_cds_sequences.fa,unblocked_cds_sequences.fa,down_cds_sequences.fa -t  2954_up,1598_unblocked,433_down -o ${home_dir}/a21-tAI-cAI/c6-tAI -u 0 -d 500 --table 1 -N
        /home/obs/${tRNA_GCNs}  && /root/miniconda3/bin/tAIPlot -i ${home_dir}/a21-tAI-cAI/c6-tAI_tAI_dataframe.txt -o ${home_dir}/a21-tAI-cAI/c7-tAIPlot -u 0 -d 500 --mode all --start 5 --window 7 --step 1  
# Local codon adaptation index and global codon adaptation index
 commands:
      - >-
        mkdir -p ${home_dir}/a21-tAI-cAI/ && cd /home/obs/${transcript_location} && /root/miniconda3/bin/cAI -i up_cds_sequences.fa,unblocked_cds_sequences.fa,down_cds_sequences.fa -o ${home_dir}/a21-tAI-cAI/c8-cAI -t 2954_up,1598_unblocked,433_down -u 0 -d 500 --reference /home/obs/${longest_cds_fa}  && 
        /root/miniconda3/bin/cAIPlot -i ${home_dir}/a21-tAI-cAI/c8-cAI_local_cAI_dataframe.txt -o ${home_dir}/a21-tAI-cAI/c9-cAIPlot -u 0 -d 500 --mode all --start 5 --window 7 --step 1
# Hydrophobicity calculation
    commands:
      - >-
      mkdir -p ${home_dir}/a22-hydropath/ && cd /home/obs/${transcript_location} &&
      /root/miniconda3/bin/hydropathyCharge  -i up_cds_sequences.fa,unblocked_cds_sequences.fa,down_cds_sequences.fa -t 2954_up,1598_unblocked,433_down -o ${home_dir}/a22-hydropath/d1-hydropathyCharge --index /home/obs/${obs_reference_hydropath} -u 0 -d 500  && 
      /root/miniconda3/bin/PlotHydropathyCharge -i ${home_dir}/a22-hydropath/d1-hydropathyCharge_values_dataframe.txt -o ${home_dir}/a22-hydropath/d3-PlotHydropathyCharge  -u 0 -d 500 --mode all --ylab "Average Hydrophobicity" 

#  Charge amino acids
    commands:
      - >-
      mkdir -p ${home_dir}/a22-hydropath/ && cd /home/obs/${transcript_location} &&
      /root/miniconda3/bin/hydropathyCharge -i up_cds_sequences.fa,unblocked_cds_sequences.fa,down_cds_sequences.fa -t 2954_up,1598_unblocked,433_down -o ${home_dir}/a22-hydropath/d2-charge --index /home/obs/${obs_reference_AAindex} -u 0 -d 500  && 
      /root/miniconda3/bin/PlotHydropathyCharge -i ${home_dir}/a22-hydropath/d2-charge_values_dataframe.txt -o ${home_dir}/a22-hydropath/d4-Plotcharges -u 0 -d 500 --mode all --ylab "Average Charges"
```

ribominer-feature-analysis-fajin
```sh
##########################
# Feature Analysis (FA)
##########################        
# RiboDensityForSpecificRegion; RiboDensityAtEachKindAAOrCodon; PlotRiboDensityAtEachKindAAOrCodon
    commands:
      - >-
        mkdir -p /home/obs/${obs_dir}/a16-ribodensitycodon2 && cd /home/obs/fjdata/data &&  /root/miniconda3/bin/RiboDensityForSpecificRegion -f /home/obs/${attributes} -c /home/obs/${longest_tra} -o /home/obs/${obs_dir}/a16-ribodensitycodon2/b5-transcript-enrich -U codon -M RPKM -L 25 -R 75  &&
        /root/miniconda3/bin/RiboDensityAtEachKindAAOrCodon -f /home/obs/${attributes} -c /home/obs/${longest_tra} -o /home/obs/${obs_dir}/a16-ribodensitycodon2/b6-ribosome-aa -M counts  -l 100 -n 10 --table 1 -F /home/obs/${trans_cds_seq} &&  /root/miniconda3/bin/PlotRiboDensityAtEachKindAAOrCodon -i /home/obs/${obs_dir}/a16-ribodensitycodon2/b6-ribosome-aa_all_codon_density.txt -o /home/obs/${obs_dir}/a16-ribodensitycodon2/b7-PlotRiboDensityAtEachKindAAOrCodon -g ${gname} -r
        ${rname} --level AA

# PausingScore
    commands:
   ·      - >-
        mkdir -p /home/obs/${obs_dir}/a17-PausingScore2 && cd /home/obs/fjdata/data && /root/miniconda3/bin/PausingScore -f /home/obs/${attributes} -c /home/obs/${longest_tra} -o  /home/obs/${obs_dir}/a17-PausingScore2/all -M counts -l 100 -n 10 --table 1 -F  /home/obs/${trans_cds_seq} && cd /home/obs/${obs_dir}/a17-PausingScore2/ && /root/miniconda3/bin/ProcessPausingScore -i all_si-Ctrl-1_pausing_score.txt,all_si-Ctrl-2_pausing_score.txt,all_si-eIF5A-1_pausing_score.txt,all_si-eIF5A-2_pausing_score.txt -o /home/obs/${obs_dir}/a17-PausingScore2/b9-ProcessPausingScore -g ${gname} -r ${rname} --mode raw --ratio_filter 0 --pausing_score_filter 0

# Ribosome density around the triplete amino acid (tri-AA) motifs.
#
    commands:
      - >-
      mkdir -p /home/obs/${obs_dir}/a18-AroundTriplete/ &&  cd /home/obs/fjdata/data &&
      /root/miniconda3/bin/RiboDensityAroundTripleteAAMotifs -f /home/obs/${attributes} -c /home/obs/${longest_tra} -o /home/obs/${obs_dir}/a18-AroundTriplete/c0-RiboDensityAroundTripleteAAMotifs_PPP -M counts -l 100 -n 10 --table 1 -F /home/obs/${trans_cds_seq} --type2 PPP --type1 PP && 
      /root/miniconda3/bin/PlotRiboDensityAroundTriAAMotifs -i /home/obs/${obs_dir}/a18-AroundTriplete/c0-RiboDensityAroundTripleteAAMotifs_PPP_motifDensity_dataframe.txt -o /home/obs/${obs_dir}/a18-AroundTriplete/c1-PPP_plot -g ${gname} -r ${rname} --mode mean    

# RPFdist calculation. GCContent
          commands:
      - >-
      mkdir -p ${home_dir}/a20-RPFdist-GCContent/ &&  cd /home/obs/fjdata/data &&
          /root/miniconda3/bin/RPFdist -f ${home_dir}/a9-metaplots/attributes.txt -c /home/obs/${longest_tra} -o ${home_dir}/a20-RPFdist-GCContent/c3-RPFdist -M counts -l 100 -n 10 -m 1 -e 5 &&
          /root/miniconda3/bin/GCContent -i /home/obs/${trans_cds_seq} -o ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-normal --mode normal &&
          /root/miniconda3/bin/GCContent -i /home/obs/${trans_cds_seq} -o ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-frames --mode frames &&
          /root/miniconda3/bin/PlotGCContent -i ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-normal_GC_content.txt -o ${home_dir}/a20-RPFdist-GCContent/c5-PlotGCContent-normal --mode normal &&
          /root/miniconda3/bin/PlotGCContent -i ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-frames_GC_content_frames.txt -o ${home_dir}/a20-RPFdist-GCContent/c5-PlotGCContent-frames --mode frames

# Local tRNA adaptation index and global tRNA adaptation index
    commands:
      - >-  
        mkdir -p ${home_dir}/a21-tAI-cAI/ && cd /home/obs/${transcript_location} && /root/miniconda3/bin/tAI -i up_cds_sequences.fa,unblocked_cds_sequences.fa,down_cds_sequences.fa -t  2954_up,1598_unblocked,433_down -o ${home_dir}/a21-tAI-cAI/c6-tAI -u 0 -d 500 --table 1 -N
        /home/obs/${tRNA_GCNs}  && /root/miniconda3/bin/tAIPlot -i ${home_dir}/a21-tAI-cAI/c6-tAI_tAI_dataframe.txt -o ${home_dir}/a21-tAI-cAI/c7-tAIPlot -u 0 -d 500 --mode all --start 5 --window 7 --step 1  
# Local codon adaptation index and global codon adaptation index
 commands:
      - >-
        mkdir -p ${home_dir}/a21-tAI-cAI/ && cd /home/obs/${transcript_location} && /root/miniconda3/bin/cAI -i up_cds_sequences.fa,unblocked_cds_sequences.fa,down_cds_sequences.fa -o ${home_dir}/a21-tAI-cAI/c8-cAI -t 2954_up,1598_unblocked,433_down -u 0 -d 500 --reference /home/obs/${longest_cds_fa}  && 
        /root/miniconda3/bin/cAIPlot -i ${home_dir}/a21-tAI-cAI/c8-cAI_local_cAI_dataframe.txt -o ${home_dir}/a21-tAI-cAI/c9-cAIPlot -u 0 -d 500 --mode all --start 5 --window 7 --step 1
# Hydrophobicity calculation
    commands:
      - >-
      mkdir -p ${home_dir}/a22-hydropath/ && cd /home/obs/${transcript_location} &&
      /root/miniconda3/bin/hydropathyCharge  -i up_cds_sequences.fa,unblocked_cds_sequences.fa,down_cds_sequences.fa -t 2954_up,1598_unblocked,433_down -o ${home_dir}/a22-hydropath/d1-hydropathyCharge --index /home/obs/${obs_reference_hydropath} -u 0 -d 500  && 
      /root/miniconda3/bin/PlotHydropathyCharge -i ${home_dir}/a22-hydropath/d1-hydropathyCharge_values_dataframe.txt -o ${home_dir}/a22-hydropath/d3-PlotHydropathyCharge  -u 0 -d 500 --mode all --ylab "Average Hydrophobicity" 

#  Charge amino acids
    commands:
      - >-
      mkdir -p ${home_dir}/a22-hydropath/ && cd /home/obs/${transcript_location} &&
      /root/miniconda3/bin/hydropathyCharge -i up_cds_sequences.fa,unblocked_cds_sequences.fa,down_cds_sequences.fa -t 2954_up,1598_unblocked,433_down -o ${home_dir}/a22-hydropath/d2-charge --index /home/obs/${obs_reference_AAindex} -u 0 -d 500  && 
      /root/miniconda3/bin/PlotHydropathyCharge -i ${home_dir}/a22-hydropath/d2-charge_values_dataframe.txt -o ${home_dir}/a22-hydropath/d4-Plotcharges -u 0 -d 500 --mode all --ylab "Average Charges"
```

Enrichment Analysis (EA)
```sh
# Step 1: Calculate ribosome density at each position for each transcript.
    commands:
      - >-
      mkdir -p ${home_dir}/EnrichmentAnalysis/GUS && mkdir -p ${home_dir}/EnrichmentAnalysis/MES && cd /home/obs/fjdata/data &&  
      /root/miniconda3/bin/RiboDensityAtEachPosition -c /home/obs/${longest_tra} -f /home/obs/${GUS_attributes} -o ${home_dir}/EnrichmentAnalysis/GUS/GUS  -U codon  && 
      /root/miniconda3/bin/RiboDensityAtEachPosition -c /home/obs/${longest_tra} -f /home/obs/${MES_attributes} -o ${home_dir}/EnrichmentAnalysis/MES/MES  -U codon 

# Step 2: Calculate mean ribosome density for different replicates.
    commands:
      - >-
        /root/miniconda3/bin/enrichmentMeanDensity -i ${home_dir}/EnrichmentAnalysis/GUS/GUS_GUS1-IP-1_cds_codon_density.txt,${home_dir}/EnrichmentAnalysis/GUS/GUS_GUS1-IP-2_cds_codon_density.txt -o ${home_dir}/EnrichmentAnalysis/GUS/GUS_IP && 
        /root/miniconda3/bin/enrichmentMeanDensity -i ${home_dir}/EnrichmentAnalysis/GUS/GUS_GUS1-total-1_cds_codon_density.txt,${home_dir}/EnrichmentAnalysis/GUS/GUS_GUS1-total-2_cds_codon_density.txt -o ${home_dir}/EnrichmentAnalysis/GUS/GUS_total && 
        /root/miniconda3/bin/enrichmentMeanDensity -i ${home_dir}/EnrichmentAnalysis/MES/MES_MES1-IP-1_cds_codon_density.txt,${home_dir}/EnrichmentAnalysis/MES/MES_MES1-IP-2_cds_codon_density.txt -o ${home_dir}/EnrichmentAnalysis/MES/MES_IP && 
        /root/miniconda3/bin/enrichmentMeanDensity -i ${home_dir}/EnrichmentAnalysis/MES/MES_MES1-total-1_cds_codon_density.txt,${home_dir}/EnrichmentAnalysis/MES/MES_MES1-total-2_cds_codon_density.txt -o ${home_dir}/EnrichmentAnalysis/MES/MES_total


# Step 3: Enrichment analysis.
    commands:
      - >-
        /root/miniconda3/bin/EnrichmentAnalysis --ctrl ${home_dir}/EnrichmentAnalysis/GUS/GUS_total_mean_density.txt --treat ${home_dir}/EnrichmentAnalysis/GUS/GUS_IP_mean_density.txt -c /home/obs/${longest_tra} -o ${home_dir}/EnrichmentAnalysis/GUS/GUS -U codon -M RPKM -l 150 -n 10 -m 1 -e 30 --CI 0.95 -u 0 -d 500  &&  
        /root/miniconda3/bin/EnrichmentAnalysis --ctrl ${home_dir}/EnrichmentAnalysis/MES/MES_total_mean_density.txt --treat ${home_dir}/EnrichmentAnalysis/MES/MES_IP_mean_density.txt -c /home/obs/${longest_tra} -o ${home_dir}/EnrichmentAnalysis/MES/MES -U codon -M RPKM -l 150 -n 10 -m 1 -e 30 --CI 0.95 -u 0 -d 500  

# Step 4: Plot the enrichment ratio.


# Notes: if you want to see the enrichment ratio for a single transcript, the EnrichmentAnalysisForSingleTrans would be helpful.
    commands:
      - >-
      /root/miniconda3/bin/EnrichmentAnalysisForSingleTrans -i ${home_dir}/EnrichmentAnalysis/MES/MES_codon_ratio.txt -s GUS1 -o ${home_dir}/EnrichmentAnalysis/MES/MES_GUS1 -c /home/obs/${longest_tra}  --id-type gene_name --slide-window y --axhline 1 && 
      /root/miniconda3/bin/EnrichmentAnalysisForSingleTrans -i ${home_dir}/EnrichmentAnalysis/MES/MES_codon_ratio.txt -s ARC1 -o ${home_dir}/EnrichmentAnalysis/MES/MES_ARC1 -c /home/obs/${longest_tra}  --id-type gene_name --slide-window y --axhline 1 && 
      /root/miniconda3/bin/EnrichmentAnalysisForSingleTrans -i ${home_dir}/EnrichmentAnalysis/GUS/GUS_codon_ratio.txt -s MES1 -o ${home_dir}/EnrichmentAnalysis/GUS/GUS_MES1 -c /home/obs/${longest_tra}  --id-type gene_name --slide-window y --axhline 1 && 
      /root/miniconda3/bin/EnrichmentAnalysisForSingleTrans -i ${home_dir}/EnrichmentAnalysis/GUS/GUS_codon_ratio.txt -s ARC1 -o ${home_dir}/EnrichmentAnalysis/GUS/GUS_ARC1 -c /home/obs/${longest_tra}  --id-type gene_name --slide-window y --axhline 1 

```

###  拟南芥没有问题。
### version 20200615
```sh
# annotation.
    commands:
      - >-
      mkdir -p /home/sfs/${JobName}/a7-RiboCode_annot &&  /root/miniconda3/bin/prepare_transcripts -g /home/obs/${obs_reference_gtf} -f /home/obs/${obs_reference_fasta} -o /home/sfs/${JobName}/a7-RiboCode_annot
#a1 annotation
    commands:
      - >-
      mkdir -p /home/sfs/${JobName}/a8-Ribominer_annot &&   /root/miniconda3/bin/OutputTranscriptInfo -c /home/sfs/${JobName}/a7-RiboCode_annot/transcripts_cds.txt -g /home/obs/${obs_reference_gtf} -f /home/sfs/${JobName}/a7-RiboCode_annot/transcripts_sequence.fa -o /home/sfs/${JobName}/a8-Ribominer_annot/longest.transcripts.info.txt -O /home/sfs/${JobName}/a8-Ribominer_annot/all.transcripts.info.txt

#a2 transcript
    commands:
      - >-
/root/miniconda3/bin/GetProteinCodingSequence -i /home/sfs/${JobName}/a7-RiboCode_annot/transcripts_sequence.fa  -c /home/sfs/${JobName}/a8-Ribominer_annot/longest.transcripts.info.txt -o /home/sfs/${JobName}/a8-Ribominer_annot/transcript --mode whole --table 1 
          commands:
      - >-

#a3 utr
    commands:
      - >-
/root/miniconda3/bin/GetUTRSequences -i /home/sfs/${JobName}/a8-Ribominer_annot/transcript_transcript_sequences.fa -o /home/sfs/${JobName}/a8-Ribominer_annot/utr -c /home/sfs/${JobName}/a7-RiboCode_annot/transcripts_cds.txt

#a4 metaplot
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a9-metaplots && /root/miniconda3/bin/metaplots -a /home/sfs/${JobName}/a7-RiboCode_annot -r ${bam_files}/${1}Aligned.toTranscriptome.out.bam -o /home/sfs/${JobName}/a9-metaplots/${1}
      vars_iter:
        - '${fastq_files}'

# sort and index.
    commands_iter:
      command: |
        samtools sort -T ${bam_files}/${1}Aligned.toTranscriptome.tmp.bam -o ${bam_files}/${1}Aligned.toTranscriptome.out.sorted.bam ${bam_files}/${1}Aligned.toTranscriptome.out.bam && samtools index ${bam_files}/${1}Aligned.toTranscriptome.out.sorted.bam && samtools index ${bam_files}/${1}Aligned.sortedByCoord.out.bam 
      vars_iter:
        - '${fastq_files}'     


#a5 periodicity
# 是的是的，需要index bam
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a10-periodicity && /root/miniconda3/bin/Periodicity -i ${bam_files}/${1}Aligned.toTranscriptome.out.sorted.bam -a /home/sfs/${JobName}/a7-RiboCode_annot -o /home/sfs/${JobName}/a10-periodicity/${1}_periodicity -c /home/sfs/${JobName}/a8-Ribominer_annot/longest.transcripts.info.txt -L 25 -R 35
      vars_iter:
        - '${fastq_files}'

#ribominer-part2
#a6 ribodensitydiffrance
    commands:
      - >-
        cd ${home_dir}/a9-metaplots/ && echo -e "#SampleName\tAlignmentFile\tStranded\tReadLength\tP-site" > attributes.txt && for i in `ls |grep _pre_config.txt`;do echo $i;grep -v "#" ${i}|grep .>> attributes.txt ;done && awk 'BEGIN {FS="\t"; OFS="\t"} {print $2, $4, $5, $1}' attributes.txt > 
        attributes2.txt && mv  attributes2.txt  attributes.txt && sed -i 's/Aligned.toTranscriptome.out.bam/Aligned.toTranscriptome.out.sorted.bam/g' attributes.txt ;
#ribominer-part3
# ribodensityofdiffframes.
    commands:
      - 'mkdir -p ${home_dir}/a11-ribodensity && /root/miniconda3/bin/RiboDensityOfDiffFrames -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a11-ribodensity/a6-ribo-density-diff-frame'
    
#ribominer-part4
#a8 dan contamination
    commands_iter:
      command: |
        mkdir -p ${home_dir}/a12-dna-contamination && /root/miniconda3/bin/StatisticReadsOnDNAsContam -i  ${bam_files}/${1}Aligned.sortedByCoord.out.bam  -g /home/obs/${obs_reference_gtf} -o  ${home_dir}/a12-dna-contamination/${1}
      vars_iter:
        - '${fastq_files}'

#ribominer-part5
# a8 metagene
    commands:
          - >-
            mkdir -p  ${home_dir}/a13-metagene && /root/miniconda3/bin/MetageneAnalysisForTheWholeRegions -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a13-metagene/a8-metagene -b 15,90,60 -l 100 -n 10 -m 1 -e 5 --plot yes && /root/miniconda3/bin/PlotMetageneAnalysisForTheWholeRegions -i ${home_dir}/a13-metagene/a8-metagene_scaled_density_dataframe.txt -o ${home_dir}/a13-metagene/a9-meta_gene_whole_regin -g ${gname} -r ${rname} -b 15,90,60 --mode all 

#ribominer-part6
#b1 meatgene
#b2 metagene utr
    commands:
        - >-
          mkdir -p ${home_dir}/a14-metageneAnalysis && /root/miniconda3/bin/MetageneAnalysis -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a14-metageneAnalysis/b1-meat-cds -U codon -M RPKM -u 0 -d 500 -l 100 -n 10 -m 1 -e 5 --norm yes -y 100 --CI 0.95 --type CDS && /root/miniconda3/bin/MetageneAnalysis -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a14-metageneAnalysis/b2-meat-utr -U nt -M RPKM -u 100 -d 100 -l 100 -n 10 -m 1 -e 5 --norm yes -y 50 --CI 0.95 --type UTR



# ribominer-part7
# b3 polarity calculation
    commands:
        - >-      
        mkdir -p ${home_dir}/a15-polarity && /root/miniconda3/bin/PolarityCalculation -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a15-polarity/b3-polarity -n 64 && 
        /root/miniconda3/bin/PlotPolarity -i ${home_dir}/a15-polarity/b3-polarity_polarity_dataframe.txt -o ${home_dir}/a15-polarity/b4-plotpolarity -g ${gname} -r ${rname} -y 5 

# ribominer-part8
#b5 transcript enrich 
#b6 ribosome aa
#b7 plot ribodensity at each aa or codon
          commands:
      - >-
      mkdir -p ${home_dir}/a16-ribodensitycodon &&  
      /root/miniconda3/bin/RiboDensityForSpecificRegion -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a16-ribodensitycodon/b5-transcript-enrich -U codon -M RPKM -L 25 -R 75 && 
      cut -f 1 ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt |sed '1d'  > ${home_dir}/a8-Ribominer_annot/select_trans.txt && /root/miniconda3/bin/RiboDensityAtEachKindAAOrCodon -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a16-ribodensitycodon/b6-ribosome-aa -M counts -S ${home_dir}/a8-Ribominer_annot/select_trans.txt  -l 100 -n 10 --table 1 -F ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa && 
      /root/miniconda3/bin/PlotRiboDensityAtEachKindAAOrCodon -i ${home_dir}/a16-ribodensitycodon/b6-ribosome-aa_all_codon_density.txt -o ${home_dir}/a16-ribodensitycodon/b7-PlotRiboDensityAtEachKindAAOrCodon -g ${gname} -r ${rname} --level AA

# ribominer-part9
#b8 pausingscore
#b9 processing pausingscore

          commands:
      - >-
       mkdir -p ${home_dir}/a17-PausingScore && 
      /root/miniconda3/bin/PausingScore -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o  ${home_dir}/a17-PausingScore/b8-PausingScore -M counts -S ${home_dir}/a8-Ribominer_annot/select_trans.txt  -l 100 -n 10 --table 1 -F  ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa && 
      /root/miniconda3/bin/ProcessPausingScore -i ${pause_name} -o ${home_dir}/a17-PausingScore/b9-ProcessPausingScore -g ${gname} -r ${rname} --mode raw --ratio_filter 2 --pausing_score_filter 0.5

# ribominer-part10
# c0 ribodenstiy around tripleaamotif
# c1 plotribodensity around tria motifs.
          commands:
      - >-
      mkdir -p ${home_dir}/a18-AroundTriplete/ && 
      /root/miniconda3/bin/RiboDensityAroundTripleteAAMotifs -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a18-AroundTriplete/c0-RiboDensityAroundTripleteAAMotifs_PPP -M counts -S ${home_dir}/a8-Ribominer_annot/select_trans.txt -l 100 -n 10 --table 1 -F ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa --type2 PPP --type1 PP && 
      /root/miniconda3/bin/PlotRiboDensityAroundTriAAMotifs -i ${home_dir}/a18-AroundTriplete/c0-RiboDensityAroundTripleteAAMotifs_PPP_motifDensity_dataframe.txt -o ${home_dir}/a18-AroundTriplete/c1-PPP_plot -g ${gname} -r ${rname} --mode mean

# ribominer-part11
# c2 ribodensity around aa motifssh
# /Users/xugang/Documents/data/reference/
## 这个是用户自己提供的，比如说之前你找到的那些可能富集更多核糖体的motif，需要自己构建
#c2b plot ribo density around aa motifs
          commands:
      - >-
      mkdir -p ${home_dir}/a19-AroundTripleteAAMotifs/ && 
      /root/miniconda3/bin/RiboDensityAroundTripleteAAMotifs -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o  ${home_dir}/a19-AroundTripleteAAMotifs/c2-RiboDensityAroundTripleteAAMotifs -M counts -S ${home_dir}/a8-Ribominer_annot/select_trans.txt -l 100 -n 10 --table 1 -F ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa --motifList1 ${home_dir}/ribominer-index/tri_AA_motifs1.txt --motifList2 ${home_dir}/ribominer-index/tri_AA_motifs2.txt && 
      /root/miniconda3/bin/PlotRiboDensityAroundTriAAMotifs -i ${home_dir}/a19-AroundTripleteAAMotifs/c2-RiboDensityAroundTripleteAAMotifs_motifDensity_dataframe.txt -o ${home_dir}/a19-AroundTripleteAAMotifs/c2b-PPP_plot -g -g ${gname} -r ${rname} --mode mean


# ribominer-part12
#c3 rpf dist
# c4 gcc
# c5 plot gcc
## normal mode
## frames mode
          commands:
      - >-
      mkdir -p ${home_dir}/a20-RPFdist-GCContent/ && 
          /root/miniconda3/bin/RPFdist -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a20-RPFdist-GCContent/c3-RPFdist -M counts -S ${home_dir}/a8-Ribominer_annot/select_trans.txt -l 100 -n 10 -m 1 -e 5 &&
          /root/miniconda3/bin/GCContent -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa -o ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-normal --mode normal &&
          /root/miniconda3/bin/GCContent -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa -o ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-frames --mode frames &&
          /root/miniconda3/bin/PlotGCContent -i ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-normal_GC_content.txt -o ${home_dir}/a20-RPFdist-GCContent/c5-PlotGCContent-normal --mode normal &&
          /root/miniconda3/bin/PlotGCContent -i ${home_dir}/a20-RPFdist-GCContent/c4-GCContent-frames_GC_content_frames.txt -o ${home_dir}/a20-RPFdist-GCContent/c5-PlotGCContent-frames --mode frames


# ribominer-part13
# c6 tAI
#c7 plot tAI
#c8 cAI
#c9 cAI plot

          commands:
      - >-
      mkdir -p ${home_dir}/a21-tAI-cAI/ && /root/miniconda3/bin/tAI -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa -t species -o ${home_dir}/a21-tAI-cAI/c6-tAI -u 0 -d 500 --table 1 -N /home/obs/${obs_reference_tRNA_confidence}  &&
       /root/miniconda3/bin/tAIPlot -i ${home_dir}/a21-tAI-cAI/c6-tAI_tAI_dataframe.txt -o ${home_dir}/a21-tAI-cAI/c7-tAIPlot -u 0 -d 500 --mode all --start 5 --window 7 --step 1 &&
       /root/miniconda3/bin/cAI -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa -o ${home_dir}/a21-tAI-cAI//c8-cAI -t tair -u 0 -d 500 --reference ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt  && 
       /root/miniconda3/bin/cAIPlot -i ${home_dir}/a21-tAI-cAI/c8-cAI_local_cAI_dataframe.txt -o ${home_dir}/a21-tAI-cAI/c9-cAIPlot -u 0 -d 500 --mode all --start 5 --window 7 --step 1

# ribominer-part14
# d1 hydropath charge
## 用户提供：hydropathy_index.txt AA_charge.txt 
# d3 plot hydropath charge
# d4 plot charges
          commands:
      - >-
      mkdir -p ${home_dir}/a22-hydropath/ && \
      /root/miniconda3/bin/hydropathyCharge  -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa -o ${home_dir}/a22-hydropath/d1-hydropathyCharge -t select_gene --index /home/obs/${obs_reference_hydropath} -u 0 -d 500 --table 1 && \
      /root/miniconda3/bin/PlotHydropathyCharge -i ${home_dir}/a22-hydropath/d1-hydropathyCharge_values_dataframe.txt -o ${home_dir}/a22-hydropath/d3-PlotHydropathyCharge  -u 0 -d 500 --mode all --ylab "Average Hydrophobicity" && \
      /root/miniconda3/bin/PlotHydropathyCharge -i ${home_dir}/a22-hydropath/d2-charge_values_dataframe.txt -o ${home_dir}/a22-hydropath/d4-Plotcharges -u 0 -d 500 --mode all --ylab "Average Charges"

# ribominer-part15
#d5 ribodensity at each postion
#d6 enrichment mean density
#d7 enrichment analysis
#d8 plot enrichment raito.

          commands:
      - >-
      mkdir -p ${home_dir}/a23-densityenrich/ && 
      /root/miniconda3/bin/RiboDensityAtEachPosition -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -f ${home_dir}/a9-metaplots/attributes.txt -o ${home_dir}/a23-densityenrich/d5-RiboDensityAtEachPosition -U codon && \
      /root/miniconda3/bin/enrichmentMeanDensity -i  ${enrichmentMeanDensity_name} -o ${home_dir}/a23-densityenrich/d6-enrichmentMeanDensity && \

      /root/miniconda3/bin/EnrichmentAnalysis --ctrl ${home_dir}/a23-densityenrich/${enrichmentMeanDensity_ctrl} --treat ${home_dir}/a23-densityenrich/${enrichmentMeanDensity_case} -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a23-densityenrich/d7-EnrichmentAnalysis -U codon -M RPKM -l 150 -n 10 -m 1 -e 30 --CI 0.95 -u 0 -d 500 && \

      /root/miniconda3/bin/PlotEnrichmentRatio -i ${home_dir}/a23-densityenrich/d7-EnrichmentAnalysis_enrichment_dataframe.txt -o ${home_dir}/a23-densityenrich/d8-PlotEnrichmentRatio -u 0 -d 500 --unit codon --mode all



```

```sh
#ao ribocode
prepare_transcripts -g /data/reference/tair/Arabidopsis_thaliana.TAIR10.43.gtf -f /data/reference/tair/Arabidopsis_thaliana.TAIR10.dna.toplevel.fa -o /data/reference/RiboCode_annot

#a1 annotation
OutputTranscriptInfo -c /data/reference/RiboCode_annot/transcripts_cds.txt -g /data/reference/tair/Arabidopsis_thaliana.TAIR10.43.gtf -f /data/reference/RiboCode_annot/transcripts_sequence.fa -o ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -O ${home_dir}/a8-Ribominer_annot/all.transcripts.info.txt

#a2 transcript
GetProteinCodingSequence -i /data/reference/RiboCode_annot/transcripts_sequence.fa  -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o ${home_dir}/a8-Ribominer_annot/transcript --mode whole --table 1 

#a3 utr
GetUTRSequences -i ${home_dir}/a8-Ribominer_annot/transcript_transcript_sequences.fa -o ${home_dir}/a8-Ribominer_annot/utr -c /data/reference/tair_ribocode/tair/transcripts_cds.txt

#a4 metaplot
metaplots -a /data/reference/RiboCode_annot -r /data/data/colAligned.toTranscriptome.out.bam -o /data/data/a4-col
metaplots -a /data/reference/RiboCode_annot -r /data/data/d14Aligned.toTranscriptome.out.bam -o /data/data/a4-d14

#a5 periodicity
# 是的是的，需要index bam
Periodicity -i /data/data/colAligned.toTranscriptome.sort.bam -a /data/reference/RiboCode_annot -o /data/data/a5-col_periodicity -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -L 25 -R 35
Periodicity -i /data/data/d14Aligned.toTranscriptome.sort.bam -a /data/reference/RiboCode_annot -o /data/data/a5-d14_periodicity -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -L 25 -R 35

#a6 ribodensitydiffrance
RiboDensityOfDiffFrames -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/a6-ribo-density-diff-frame

#a7 dan contamination
StatisticReadsOnDNAsContam -i  /data/data/colAligned.sortedByCoord.out.bam  -g /data/reference/tair/Arabidopsis_thaliana.TAIR10.43.gtf -o  /data/data/a7-dna-contamination.col 
StatisticReadsOnDNAsContam -i  /data/data/d14Aligned.sortedByCoord.out.bam  -g /data/reference/tair/Arabidopsis_thaliana.TAIR10.43.gtf -o  /data/data/a7-dna-contamination.d14  

# a8 metagene
MetageneAnalysisForTheWholeRegions -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/a8-metagene -b 15,90,60 -l 100 -n 10 -m 1 -e 5 --plot yes

# a9 plotmetagene analysis
PlotMetageneAnalysisForTheWholeRegions -i /data/data/a8-metagene_scaled_density_dataframe.txt -o /data/data/a9-meta_gene_whole_regin -g col,d14 -r col__d14 -b 15,90,60 --mode all 

#b1 meatgene
MetageneAnalysis -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/b1-meat-cds -U codon -M RPKM -u 0 -d 500 -l 100 -n 10 -m 1 -e 5 --norm yes -y 100 --CI 0.95 --type CDS

#b2 metagene utr
MetageneAnalysis -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/b2-meat-utr -U nt -M RPKM -u 100 -d 100 -l 100 -n 10 -m 1 -e 5 --norm yes -y 50 --CI 0.95 --type UTR

# b3 polarity calculation
PolarityCalculation -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/b3-polarity -n 64

#b4 polt polarity
PlotPolarity -i /data/data/b3-polarity_polarity_dataframe.txt -o /data/data/b4-plotpolarity -g col,d14 -r col__d14 -y 5 


#b5 transcript enrich 
RiboDensityForSpecificRegion -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/b5-transcript-enrich -U codon -M RPKM -L 25 -R 75

#b6 ribosome aa
## select_trans.txt是转录本ID么师兄 如果没有的话应该会先和longest做交集的
RiboDensityAtEachKindAAOrCodon -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/b6-ribosome-aa -M counts -S /data/select_trans.txt -l 100 -n 10 --table 1 -F ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa 

#b7 plot ribodensity at each aa or codon
PlotRiboDensityAtEachKindAAOrCodon -i /data/data/b6-ribosome-aa_all_codon_density.txt -o /data/data/b7-PlotRiboDensityAtEachKindAAOrCodon -g col,d14 -r col__d14 --level AA

#b8 pausingscore
PausingScore -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o  /data/data/b8-PausingScore -M counts -S /data/select_trans.txt  -l 100 -n 10 --table 1 -F  ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa

#b9 processing pausingscore
ProcessPausingScore -i /data/data/b8-PausingScore_col_pausing_score.txt,/data/data/b8-PausingScore_d14_pausing_score.txt -o /data/data/b9-ProcessPausingScore -g col,d14 -r col__d14 --mode raw --ratio_filter 2 --pausing_score_filter 0.5

# c0 ribodenstiy around tripleaamotif
RiboDensityAroundTripleteAAMotifs -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/c0-RiboDensityAroundTripleteAAMotifs_PPP -M counts -S /data/select_trans.txt -l 100 -n 10 --table 1 -F ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa --type2 PPP --type1 PP

# c1 plotribodensity around tria motifs.
PlotRiboDensityAroundTriAAMotifs -i /data/data/c0-RiboDensityAroundTripleteAAMotifs_PPP_motifDensity_dataframe.txt -o /data/data/c1-PPP_plot -g col,d14 -r col__d14 --mode mean

# c2 ribodensity around aa motifssh
## 这个是用户自己提供的，比如说之前你找到的那些可能富集更多核糖体的motif，需要自己构建
RiboDensityAroundTripleteAAMotifs -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o  /data/data/c2-RiboDensityAroundTripleteAAMotifs -M counts -S /data/select_trans.txt -l 100 -n 10 --table 1 -F ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa --motifList1 /data/reference/tri_AA_motifs1.txt --motifList2 /data/reference/tri_AA_motifs2.txt

#c2b plot ribo density around aa motifs
PlotRiboDensityAroundTriAAMotifs -i /data/data/c2-RiboDensityAroundTripleteAAMotifs_motifDensity_dataframe.txt -o /data/data/c2b-PPP_plot -g col,d14 -r col__d14 --mode mean

#c3 rpf dist
RPFdist -f ${home_dir}/a9-metaplots/attributes.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/c3-RPFdist -M counts -S /data/select_trans.txt -l 100 -n 10 -m 1 -e 5

# c4 gcc
GCContent -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa -o /data/data/c4-GCContent-normal --mode normal
GCContent -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences.fa -o /data/data/c4-GCContent-frames --mode frames


# c5 plot gcc
## normal mode
PlotGCContent -i /data/data/c4-GCContent-normal_GC_content.txt -o /data/data/c5-PlotGCContent-normal --mode normal
## frames mode
PlotGCContent -i /data/data/c4-GCContent-frames_GC_content_frames.txt -o /data/data/c5-PlotGCContent-frames --mode frames


# c6 tAI
tAI -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences_tAI.fa -t tair -o /data/data/c6-tAI -u 0 -d 500 --table 1 -N /data/aratha/araTha1-tRNAs-confidence-set.out

#c7 plot tAI
tAIPlot -i /data/data/c6-tAI_tAI_dataframe.txt -o /data/data/c7-tAIPlot -u 0 -d 500 --mode all --start 5 --window 7 --step 1

#c8 cAI
cAI -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences_tAI.fa -o /data/data/c8-cAI -t tair -u 0 -d 500 --reference ${home_dir}/a8-Ribominer_annot/reference.fa

#c9 cAI plot
cAIPlot -i /data/data/c8-cAI_local_cAI_dataframe.txt -o /data/data/c9-cAIPlot -u 0 -d 500 --mode all --start 5 --window 7 --step 1

# d1 hydropath charge
## 用户提供：hydropathy_index.txt AA_charge.txt 
hydropathyCharge  -i ${home_dir}/a8-Ribominer_annot/transcript_cds_sequences_tAI.fa -o /data/data/d1-hydropathyCharge -t select_gene --index /data/reference/hydropathy_index.txt -u 0 -d 500 --table 1

# d3 plot hydropath charge
PlotHydropathyCharge -i /data/data/d1-hydropathyCharge_values_dataframe.txt -o /data/data/d3-PlotHydropathyCharge  -u 0 -d 500 --mode all --ylab "Average Hydrophobicity"

# d4 plot charges
PlotHydropathyCharge -i /data/data/d2-charge_values_dataframe.txt -o /data/data/d4-Plotcharges -u 0 -d 500 --mode all --ylab "Average Charges"

#d5 ribodensity at each postion
RiboDensityAtEachPosition -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -f ${home_dir}/a9-metaplots/attributes.txt -o /data/data/d5-RiboDensityAtEachPosition -U codon

#d6 enrichment mean density
enrichmentMeanDensity -i /data/data/d5-RiboDensityAtEachPosition_col_cds_codon_density.txt,/data/data/d5-RiboDensityAtEachPosition_d14_cds_codon_density.txt -o /data/data/d6-enrichmentMeanDensity

#d7 enrichment analysis
EnrichmentAnalysis --ctrl /data/data/d5-RiboDensityAtEachPosition_col_cds_codon_density.txt --treat /data/data/d5-RiboDensityAtEachPosition_d14_cds_codon_density.txt -c ${home_dir}/a8-Ribominer_annot/longest.transcripts.info.txt -o /data/data/d7-EnrichmentAnalysis -U codon -M RPKM -l 150 -n 10 -m 1 -e 30 --CI 0.95 -u 0 -d 500

#d8 plot enrichment raito.
PlotEnrichmentRatio -i /data/data/d7-EnrichmentAnalysis_enrichment_dataframe.txt -o /data/data/d8-PlotEnrichmentRatio -u 0 -d 500 --unit codon --mode all
```
```sh
docker run -dt --name ribominer -v ~/Downloads/data/:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/ribominer:1.0

docker exec -it ribominer bash

exit
docker stop ribominer
docker rm ribominer

```
## 

```sh
mkdir -p /home/sfs/a5-rmrRNA && \ mkdir -p /home/sfs/a5-rmrRNA/nonrRNA && \ echo SRR3498212.fq begin `date` && \ bash /root/.bashrc && \ /home/test/bowtie-1.2.3-linux-x86_64/bowtie \ -n 0 -norc --best -l 15 -p 8 \ --un=/home/sfs/a5-rmrRNA/nonrRNA/nocontam_SRR3498212.fq /home/obs/arabidopsis/huawei_file/refrence/tair_rRNA_bowtie_index/tair.rRNA.fa \ -q /home/sfs/a3-filter/SRR3498212.fq_trimmedQfilter.fastq \ /home/sfs/a5-rmrRNA/SRR3498212.fq.alin > \ /home/sfs/a5-rmrRNA/SRR3498212.fq.err && \ rm -rf /home/sfs/a5-rmrRNA/SRR3498212.fq.alin

obsutil config -i=5ULAGR0CWKBAEDV57Y6P -k=gvroYZE9uUmp3igpEPAEQRfuQzUjcVQn9kBoHz02 -e=https://obs.cn-north-4.myhuaweicloud.com&& obsutil mkdir -p obs://hw-gcs-logo-cn-north-4-06a54be3938010610f01c00da675d700/output/arabidopsis-smallrnaseq/ && obsutil cp -r -f /home/sfs/arabidopsis-smallrnaseq/ obs://hw-gcs-logo-cn-north-4-06a54be3938010610f01c00da675d700/output/arabidopsis-smallrnaseq/ && rm -rf /home/sfs/arabidopsis-smallrnaseq && echo Check sfs && ls -al /home/sfs && ls -al /home/obs/output
```

## blastn
```sh
    commands:
      - 'mkdir -p /home/obs/${data_output} && blastn  -query /home/obs/${data_query}  -subject   /home/obs/${data_subject} -out /home/obs/${data_output}/output.txt'
    
```

## blastp
```sh
    commands:
      - 'blastp  -query /home/obs/${data_query}  -subject   /home/obs/${data_subject} -out /home/obs/${data_output}/output.txt'
    
```

## alternative splicing
```sh
          commands:
      - >-
echo "input/SRR065544_chrX.bam" > input/b1.txt && \
echo "input/SRR065545_chrX.bam" > input/b2.txt && \
python2 /usr/local/rMATS-turbo-Linux-UCS4/rmats.py \
    --b1 input/b1.txt --b2 input/b2.txt --gtf input/Mus_musculus_chrX.gtf --od output \
    -t paired --readLength 35

      commands:
      - >-
        mkdir -p /home/obs/${data_output} && echo /home/obs/${ipbam} > /home/obs/${data_output}/b1.txt && \ echo /home/obs/${bgbam} > /home/obs/${data_output}/b2.txt && \ python2 /home/app/rMATS.4.0.2/rMATS-turbo-Linux-UCS4/rmats.py \ --b1 /home/obs/${data_output}/b1.txt --b2 \
        /home/obs/${data_output}/b2.txt --gtf ${gtf} --od /home/obs/${data_output} \ -t paired --readLength 35
        
```

## Chip-seq
```sh
# remove adapter
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a2-cutadapter && \
        echo ${1} begin `date` /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        /home/obs/${obs_data_path}/${1} && \
        /root/miniconda3/bin/cutadapt -m 18 \
        --match-read-wildcards -a ${adapter} \
        -o /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
         /home/obs/${obs_data_path}/${1}
      vars_iter:
        - '${fastq_files}'
# remove low quality
   commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a3-filter && \
        echo ${1} begin `date` && \
        /home/test/bin/fastq_quality_filter \
        -Q33 -v -q 25 -p 75 \
        -i /home/sfs/${JobName}/a2-cutadapter/${1}_trimmed.fastq \
        -o /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq 
      vars_iter:
        - '${fastq_files}'
    commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a4-qc && \
        echo ${1} begin `date` && \
        /home/test/FastQC/fastqc \
        /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
        -o /home/sfs/${JobName}/a4-qc
      vars_iter:
        - '${fastq_files}'

#bowtie mapping
   commands_iter:
      command: |
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA && \
        mkdir -p /home/sfs/${JobName}/a5-rmrRNA/nonrRNA && \
        echo ${1} begin `date` && \
        bash /root/.bashrc && \
        /home/test/bowtie-1.2.3-linux-x86_64/bowtie \
        -n 0 -norc --best -l 15 -p 8 \
         --un=/home/sfs/${JobName}/a5-rmrRNA/nonrRNA/nocontam_${1} /home/obs/${obs_reference_rRNA_bowtie} \
         -q /home/sfs/${JobName}/a3-filter/${1}_trimmedQfilter.fastq \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.sam > \
         /home/sfs/${JobName}/a5-rmrRNA/${1}.err
         
      vars_iter:
        - '${fastq_files}'
        
# sam to bam
    commands_iter:
      command: |
        samtools view -bS -o /home/sfs/${JobName}/a5-rmrRNA/${1}.bam /home/sfs/${JobName}/a5-rmrRNA/${1}.sam &&
        samtools sort /home/sfs/${JobName}/a5-rmrRNA/${1}.bam /home/sfs/${JobName}/a5-rmrRNA/${1}.sorted   && 
        samtools index /home/sfs/${JobName}/a5-rmrRNA/${1}.sorted.bam
      vars_iter:
        - '${fastq_files}'

# homer peak
commands:
      - >-
        mkdir -p  /home/obs/${data_output} && /home/app/hommer/bin/makeTagDirectory /home/obs/${data_output}/ip /home/sfs/${JobName}/a5-rmrRNA/${ipbam}.sorted.bam && export PATH=/home/app/hommer/bin/:$PATH && /home/app/hommer/bin/makeTagDirectory /home/obs/${data_output}/input 
        /home/sfs/${JobName}/a5-rmrRNA/${bgbam}.sorted.bam && /home/app/hommer/bin/findPeaks /home/obs/${data_output}/ip -style factor -o /home/obs/${data_output}/part.peak -i /home/obs/${data_output}/input && /home/app/hommer/bin/findMotifsGenome.pl /home/obs/${data_output}/part.peak ${spieces}
        /home/obs/${data_output}/part.motif.output -len 8 && mv /home/sfs/${JobName} /home/obs/${data_output}

```

## bash with parameters

**merge.sh**

```sh
#!/bin/bash

helpFunction()
{
   echo ""
   echo "Usage: $0 -a FileName -b Lables -c Directory"
   echo -e "\t-a Description of what is Filename,like file1.counter;file2.counter;file3.counter;file4.counter"
   echo -e "\t-b Description of what is Label, like dark1;dark2;dark3;dark4"
   echo -e "\t-c Description of what is Diectory path:/Users/xugang/Downloads"
   echo -e '\t ./merge.sh -a "SRR3498212.fq;SRR3498213.fq;SRR966479.fq;SRR966480.fq" -b "control1;control2;case1;case2" -c "./"'
   echo "chmod 755 ./merge.sh";
   exit 1 # Exit script after printing help
 }

 while getopts "a:b:c:" opt
 do
    case "$opt" in
      a ) parameterA="$OPTARG" ;;
        b ) parameterB="$OPTARG" ;;
      c ) parameterC="$OPTARG" ;;
        ? ) helpFunction ;; # Print helpFunction in case parameter is non-existent
   esac
   done

   # Print helpFunction in case parameters are empty
   if [ -z "$parameterA" ] || [ -z "$parameterB" ] || [ -z "$parameterC" ]
   then
      echo "Some or all of the parameters are empty";
     helpFunction
 fi

 # Begin script in case all parameters are correct
 echo "$parameterA"
 echo "$parameterB"
 echo "$parameterC"

 cd $parameterC

 IFS=';' read -ra name <<< "$parameterA"
 IFS=';' read -ra label <<< "$parameterB"

 for i in "${name[@]}";
 do
 echo $i;
 done

 merge_file(){
 head='gene'
 for i in ${label[@]};
 do
  head+=" ${i}";
  done
  echo -e $head >merge.counter;

  begin1=${name[0]};
  begin2=${name[1]};
  name2=("${name[@]:2}");
  join ${begin1}.count ${begin2}.count >merge.tmp
  commander='join';
  for i in ${name2[@]};
  do
  echo ${i}.count;
  join merge.tmp ${i}.count >>merge.tmp2;
  mv merge.tmp2 merge.tmp
  done

  cat merge.counter merge.tmp > merge2.tmp;
  rm merge.tmp
  mv merge2.tmp merge.counter
  sed -i 's/ \+/\t/g' merge.counter

  grep -v '^__' merge.counter > merge.counter2
  mv merge.counter2 heat.counter
  }

merge_file

```

**test script:**

```sh
sh merge.sh -a "SRR3498212.fq;SRR3498213.fq;SRR966479.fq;SRR966480.fq" -b "control1;control2;case1;case2" -c "./"

sh ./myscript -a "riboseq_heat;RNAseq_heat;riboseq_control_heat2;RNAseq_control_heat2;riboseq_control_heat1;RNAseq_control_heat1" -b "riboseq_heat;RNAseq_heat;riboseq_control_heat2;RNAseq_control_heat2;riboseq_control_heat1;RNAseq_control_heat1" -c "./"
String A
String B
String C

$ ./myscript -a "SRR966479.fq;" -c "String C" -b "String B"
String A
String B
String C

$ ./myscript -a "String A" -c "String C" -f "Non-existent parameter"
./myscript: illegal option -- f

Usage: ./myscript -a parameterA -b parameterB -c parameterC
    -a Description of what is parameterA
    -b Description of what is parameterB
    -c Description of what is parameterC

$ ./myscript -a "String A" -c "String C"
Some or all of the parameters are empty

Usage: ./myscript -a parameterA -b parameterB -c parameterC
    -a Description of what is parameterA
    -b Description of what is parameterB
    -c Description of what is parameterC


```
```r
args <- commandArgs(trailingOnly = TRUE)
# inputdata
filename=args[1]
# ribovector
ribovector=as.integer(unlist(strsplit(args[2],",")))
# rnavector
rnavector=as.integer(unlist(strsplit(args[3],",")))
# label mean
label=unlist(strsplit(args[4],","))
# output
output=args[5]
print(filename)
print(ribovector)
print(rnavector)
print(label)
print(paste(output,"/df",sep=""))


library(xtail)
lbxd=read.table(filename,header=T,row.name=1)
mrna=lbxd[,ribovector]
rpf=lbxd[,rnavector]
condition=label
test.results=xtail(mrna,rpf,condition,bins=1000,threads=2)
summary(test.results)

#
test.tab=resultsTable(test.results);
head(test.tab,5)

write.table(test.tab,paste(output,"/control_case_results.txt",sep=""),quote=F,sep="\t");

# Visualization
pdf(paste(output,"/control_caseFC.pdf",sep=""),width=6,height=4,paper='special')
lbxfc=plotFCs(test.results)
dev.off()
write.table(lbxfc$resultsTable,paste(output,"/control_casefc_results.txt",sep=""),quote=F,sep="\t");

pdf(paste(output,"control_caseRs.pdf",sep=""),width=6,height=4,paper='special')
lbxrs=plotRs(test.results)
dev.off()
write.table(lbxrs$resultsTable,paste(output,"/control_casers_results.txt",sep=""),quote=F,sep="\t");

pdf(paste(output,"/control_casevolcano.pdf",sep=""),width=6,height=4,paper='special')
volcanoPlot(test.results)
dev.off()
```

```sh
# run
Rscript xtail.r merge2.counter 1,2 3,4 control,case ./output/
```


## 对字符串进行分割成数组

```sh
string='["SRR3498212.fq","SRR966479.fq"]'
string=`sed "s/\"//g" <<<"$string"`
string=`sed "s/\[//g" <<<"$string"`
string=`sed "s/\]//g" <<<"$string"`
IFS=', ' read -r -a array <<< "$string"

for element in "${array[@]}"
do
    echo "$element"
done
```
