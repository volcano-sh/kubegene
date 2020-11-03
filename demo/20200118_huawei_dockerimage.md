# 华为Docker image

-----------------------------

## [Huawei Homepage](20200118_huawei.md)

## [Huawei Docker images](20200118_huawei_dockerimage.md)

## [Huawei 文件拷贝](20200118_huawei_copy.md)

## [Huawei GCS 流程设计](20200118_huawei_gcspipeline.md)

-----------------------------

## Content

* [上传镜像](#上传镜像)
* [HTSeq](#HTSeq)
* [fastqc](#fastqc)
* [xtail](#xtail)
* [bowtie](#bowtie)
* [HTSeq](#HTSeq)
* [Ribocode](#Ribocode)
* [Ribominer](#Ribominer)
* [cutadapter](#cutadapter)
* [fastx_toolkit](#fastx_toolkit)
* [bedtools](#bedtools)
* [samtools](#samtools)
* [镜像的导出](#镜像的导出)


## 上传镜像

```sh
docker save gangxu/base_ubuntu:4.0 | gzip > gangxu_base_ubuntu.tar.gz

docker login -u cn-north-4@Y3NHYJC8KGOGABXQMM9H -p 70216640613c345678fdb439ce901fe4cb83546ea2b429abc13ee23a60913fa0 swr.cn-north-4.myhuaweicloud.com
docker login -u cn-north-4@IES26GBWN2NH8G4ESXEZ -p a88225c20c59a275ee2083f03ea23be53185f93a4f06760978ff03814af3dbe3 swr.cn-north-4.myhuaweicloud.com

$ sudo docker tag [{镜像名称}:{版本名称}] swr.cn-north-4.myhuaweicloud.com/{组织名称}/{镜像名称}:{版本名称}
$ sudo docker tag gangxu/base_ubuntu:4.0 swr.cn-north-4.myhuaweicloud.com/gangxu/base:1.0
$ sudo docker push swr.cn-north-4.myhuaweicloud.com/{组织名称}/{镜像名称}:{版本名称}
$ sudo docker push swr.cn-north-4.myhuaweicloud.com/gangxu/base:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/ribocode_ribominer:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/htseq:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/ribominer:1.0
docker tag swr.cn-north-4.myhuaweicloud.com/gangxu/htseq:1.0 gangxu/htseq:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/blast:1.0

docker push gangxu/htseq:1.0

yanglab
62783319d226yang

```

```sh
docker run -dt --name base -v /home/xugang/app:/home/app gangxu/base_ubuntu:4.0
docker exec -it base bash
exit
docker stop base
docker rm base
```

```sh
docker ps|grep base
[xugang@hub app]$ docker ps|grep base
d881f9fdee83        gangxu/base_ubuntu:4.0     "/bin/bash"              About a minute ago   Up About a minute                                                                    base

```

## fastqc

```sh
docker run -dt --name base -v /home/xugang/app:/home/app gangxu/base_ubuntu:4.0
docker exec -it base bash
wget https://www.bioinformatics.babraham.ac.uk/projects/fastqc/fastqc_v0.11.9.zip
apt-get update
apt-get install libcam-pdf-perl
apt-get install default-jdk
docker commit d881f9fdee83 swr.cn-north-4.myhuaweicloud.com/gangxu/fastqc:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/fastqc:1.0
docker stop base
docker rm base

docker run -dt --name fastqc -v /home/xugang/singularity_image/huawei_file:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/fastqc:1.0
docker exec -it fastqc bash
/home/test/FastQC/fastqc
exit
docker stop fastqc
docker rm fastqc
```

## xtail
```sh
docker tag xug15/xtail:latest swr.cn-north-4.myhuaweicloud.com/gangxu/xtail:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/xtail:1.0
docker run -dt --name xtail -v /home/xugang/singularity_image/huawei_file:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/cutadapter:1.0
docker exec -it xtail bash
docker stop xtail
docker rm xtail
```
```sh
sudo singularity build --sandbox /home/xugang/singularity_image/xtail docker://gangxu/xtail:latest

```

## bowtie
```sh
docker run -dt --name base -v /home/xugang/app:/home/app gangxu/base_ubuntu:4.0
docker exec -it base bash
wget https://sourceforge.net/projects/bowtie-bio/files/bowtie/1.2.3/bowtie-1.2.3-linux-x86_64.zip
wget https://sourceforge.net/projects/bowtie-bio/files/bowtie2/2.3.5.1/bowtie2-2.3.5.1-linux-x86_64.zip
wget https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh 
apt-get update
apt-get install libcam-pdf-perl
apt-get update
  153  apt-get install python
  154  apt-get --reinstall install python-minimal
  155  apt-get --reinstall install python-minimal
exit
docker commit a9a6063a86fd swr.cn-north-4.myhuaweicloud.com/gangxu/bowtie12:1.1
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/bowtie12:1.1
docker run -dt --name bowtie -v /home/xugang/singularity_image/huawei_file:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/bowtie12:1.1
/home/test/bowtie-1.2.3-linux-x86_64
/home/test/bowtie2-2.3.5.1-linux-x86_64
docker stop bowtie
docker rm bowtie

```
## HTSeq
```sh
docker run -dt --name bowtie -v /home/xugang/singularity_image/huawei_file:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/bowtie12:1.1
docker exec -it bowtie bash
/root/miniconda3/bin/htseq-count

wget https://repo.anaconda.com/miniconda/Miniconda2-latest-Linux-x86_64.sh
cd /root/miniconda2/bin
./pip install numpy
./pip install matplotlib
./pip install pysam
./pip install HTSeq
./pip install 
rm -rf /home/test/*
exit
docker ps
docker commit 1f6dea975073  swr.cn-north-4.myhuaweicloud.com/gangxu/htseq:1.0
docker stop bowtie
docker rm bowtie

docker run -dt --name htseq -v /home/xugang/singularity_image/huawei_file:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/htseq:1.0
cp merge.sh /root/miniconda2/bin
chmod 755 /root/miniconda2/bin/merge.sh
# the file merge.sh is below.
docker commit 5beb5d631ab2 swr.cn-north-4.myhuaweicloud.com/gangxu/htseq:2.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/htseq:2.0

docker stop htseq
docker rm htseq
docker run -dt --name htseq -v /home/xugang/singularity_image/huawei_file:/home/sfs gangxu/htseq:1.1

other verstion
cd /home/test/miniconda2/bin
./pip install numpy
./pip install matplotlib
./pip install pysam
./pip install HTSeq
exit
docker ps
docker commit ce6a6eeece26  gangxu/htseq:1.1
docker push gangxu/htseq:1.1
docker stop bowtie
docker rm bowtie
docker push  gangxu/htseq:1.1
sudo singularity build --sandbox /home/xugang/singularity_image/htseq docker://gangxu/htseq:1.1

sudo singularity shell -w  /home/xugang/singularity_image/htseq 
singularity exec -B /home/xugang/singularity_image/huawei_file:/home/sfs /home/xugang/singularity_image/htseq /home/test/miniconda2/bin/htseq-count
sudo singularity exec /home/xugang/singularity_image/htseq /home/test/miniconda2/bin/htseq-count


singularity exec /WORK/teaching/project/singularity_images/htseq /home/test/miniconda2/bin/htseq-count
```

## Ribocode
```sh
docker run -dt --name bowtie -v /home/xugang/singularity_image/huawei_file:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/bowtie12:1.1
docker exec -it bowtie bash

pip install ribocode
pip install RiboMiner
exit
docker commit ce6a6eeece26  swr.cn-north-4.myhuaweicloud.com/gangxu/ribocode_ribominer:1.0
docker run -dt --name ribocode -v /home/xugang/singularity_image/huawei_file:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/ribocode_ribominer:1.0
```

## Ribominer
```sh
docker run -dt --name ribominer -v ~/Downloads/data/:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/ribocode_ribominer:1.0
docker exec -it ribominer bash

pip install ribominer --upgrade
exit
#docker commit 6b77c3f83799  swr.cn-north-4.myhuaweicloud.com/gangxu/ribominer:1.0
docker commit 1ea96f908a46  swr.cn-north-4.myhuaweicloud.com/gangxu/ribominer:1.1
docker push  swr.cn-north-4.myhuaweicloud.com/gangxu/ribominer:1.1
```

## cutadapter
```sh
# use bowtie images
conda install -c bioconda cutadapt
rm -rf /home/test/bowtie*
docker commit 4790821a127b swr.cn-north-4.myhuaweicloud.com/gangxu/cutadapter:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/cutadapter:1.0
docker stop base
docker rm base

docker run -dt --name cutadapter -v /home/xugang/singularity_image/huawei_file:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/cutadapter:1.0

docker stop cutadapter
docker rm cutadapter
```

cutadapter test:

```sh
docker run -dt --name cutadapt -v /home/xugang/singularity_image/huawei_file:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/cutadapter:1.0

docker exec -it cutadapt bash
/root/miniconda3/bin/cutadapt -m 18 --match-read-wildcards -a CTGTAGGCACCATCAAT -o /home/sfs/a2-cutadapter/SRR3498212.fq_trimmed.fastq /home/sfs/a1-fastq/SRR3498212.fq 

docker stop cutadapt
docker rm cutadapt

```

## fastx_toolkit
```sh
docker run -dt --name base -v /home/xugang/app:/home/app gangxu/base_ubuntu:4.0
docker exec -it base bash
wget http://hannonlab.cshl.edu/fastx_toolkit/fastx_toolkit_0.0.13_binaries_Linux_2.6_amd64.tar.bz2
exit
docker commit 2c9d0b2e56e6 swr.cn-north-4.myhuaweicloud.com/gangxu/fastx_toolkit:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/fastx_toolkit:1.0
docker stop base
docker rm base

docker run -dt --name fastx -v /home/xugang/singularity_image:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/fastx_toolkit:1.0
docker exec -it fastx bash
/home/test/bin/fastq_quality_filter
exit

```

## bedtools
```sh
docker run -dt --name base  /home/xugang/app:/home/app gangxu/base_ubuntu:4.0
docker exec -it base bash
apt-get update
sudo apt-get install libboost-all-dev
sudo apt-get install libbz2-dev
apt-get install liblzma-dev
$ wget https://github.com/arq5x/bedtools2/releases/download/v2.29.1/bedtools-2.29.1.tar.gz
$ tar -zxvf bedtools-2.29.1.tar.gz
$ cd bedtools2
$ make
exit
docker ps|grep base
docker commit ebe6a9bfece9 swr.cn-north-4.myhuaweicloud.com/gangxu/bedtools:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/bedtools:1.0
docker stop base
docker rm base
docker run -dt --name bedtools -v /home/xugang/singularity_image:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/bedtools:1.0
docker exec -it bedtools bash



```

## samtools
```sh
docker run -dt --name base -v /home/xugang/app:/home/app gangxu/base_ubuntu:4.0
docker exec -it base bash
wget https://sourceforge.net/projects/samtools/files/samtools/1.10/samtools-1.10.tar.bz2
wget https://sourceforge.net/projects/samtools/files/samtools/1.10/bcftools-1.10.tar.bz2
apt-get update
apt-get install libncurses5-dev
apt-get install libbz2-dev
apt-get install liblzma-dev
./configure
make
make install
exit
docker ps|grep base
docker commit fc5552888cdc swr.cn-north-4.myhuaweicloud.com/gangxu/samtools2bcftools:1.0
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/samtools2bcftools:1.0
docker stop base
docker rm base
docker run -dt --name samtools -v /home/xugang/singularity_image:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/samtools2bcftools:1.0
samtools exec -it samtools bash
```

## blast
```sh
docker exec -it ribocode_ribominer bash
sudo apt-get install ncbi-blast+
# 

docker run -dt --name ribocode_ribominer yanglab/ribocode_ribominer:
docker exec -it ribocode_ribominer bash
apt-get install ncbi-blast+
exit
docker commit 6fc91d01a4f8 swr.cn-north-4.myhuaweicloud.com/gangxu/blast:1.1

```

## Chip-seq
```sh
 docker exec -it ribocode_ribominer bash


wget http://homer.ucsd.edu/homer/configureHomer.pl
apt-get install zip unzip
perl configureHomer.pl -install homer
docker commit 6fc91d01a4f8 swr.cn-north-4.myhuaweicloud.com/gangxu/hommer:1.1
docker push  swr.cn-north-4.myhuaweicloud.com/gangxu/hommer:1.1
docker run --name chipseq -dt -v /Users/xugang/Desktop/bioinfo_tsinghua_share:/home/sfs swr.cn-north-4.myhuaweicloud.com/gangxu/hommer:1.1
docker exec -it chipseq bash

wget https://sourceforge.net/projects/samtools/files/samtools/1.10/samtools-1.10.tar.bz2
wget https://sourceforge.net/projects/samtools/files/samtools/1.10/bcftools-1.10.tar.bz2
apt-get update
apt-get install libncurses5-dev
apt-get install libbz2-dev
apt-get install liblzma-dev

docker commit 1615d7bbb1c4 swr.cn-north-4.myhuaweicloud.com/gangxu/hommer:1.2
docker push  swr.cn-north-4.myhuaweicloud.com/gangxu/hommer:1.2


```

## alternative splicing
```sh

docker exec -it ribocode_ribominer bash

# For Ubuntu 14:
pip install numpy
apt-get update
sudo apt-get install libblas-dev liblapack-dev
sudo apt-get install libgsl0ldbl
apt-get install libgsl23 libgslcblas0
sudo apt-get install gfortran

http://rnaseq-mats.sourceforge.net/rmats4.0.2/user_guide.htm

tar -xzf rMATS.4.0.2.tgz
cd rMATS.4.0.2/
... # move/copy/download data to this folder.
tar -xzf gtf.tgz
tar -xzf testData.tgz

cd rMATS.4.0.2/
python2 rMATS-turbo-xxx-UCSx/rmats.py --b1 b1.txt --b2 b2.txt --gtf gtf/Homo_sapiens.Ensembl.GRCh37.75.gtf --od bam_test -t paired --readLength 50 --cstat 0.0001 --libType fr-unstranded

apt-get install python-numpy python-scipy
apt-get install libgfortran3

docker commit 6fc91d01a4f8 swr.cn-north-4.myhuaweicloud.com/gangxu/alternative:1.2
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/alternative:1.2


docker pull nunoagostinho/rmats
docker run --name rmats -dt -v /home/xugang:/home/sfs nunoagostinho/rmats
docker exec -it rmats bash
wget https://github.com/Xinglab/rmats-turbo/releases/download/v4.1.0/rmats_turbo_v4_1_0.tar.gz
cd /root/software/rMATS.3.2.5  && python RNASeq-MATS.py -b1 testData/231ESRP.25K.rep-1.bam,testData/231ESRP.25K.rep-2.bam -b2 testData/231EV.25K.rep-1.bam,testData/231EV.25K.rep-2.bam -gtf testData/test.gtf -o bam_test -t paired -len 50 -a 8 -c 0.0001 -analysis U -novelSS 1 -keepTemp

docker tag  nunoagostinho/rmats swr.cn-north-4.myhuaweicloud.com/gangxu/alternative:1.3
docker push swr.cn-north-4.myhuaweicloud.com/gangxu/alternative:1.3

```

## 镜像的导出

```sh
docker save swr.cn-north-4.myhuaweicloud.com/gangxu/bowtie12:1.1 | gzip > /home/xugang/singularity_image/huawei_file/images/bowtie12.tar.gz
docker save swr.cn-north-4.myhuaweicloud.com/gangxu/samtools2bcftools:1.0 | gzip > /home/xugang/singularity_image/huawei_file/images/samtools2bcftools.tar.gz
docker save swr.cn-north-4.myhuaweicloud.com/gangxu/bedtools:1.0 | gzip > /home/xugang/singularity_image/huawei_file/images/bedtools.tar.gz
docker save swr.cn-north-4.myhuaweicloud.com/gangxu/fastx_toolkit:1.0 | gzip > /home/xugang/singularity_image/huawei_file/images/fastx_toolkit.tar.gz
docker save swr.cn-north-4.myhuaweicloud.com/gangxu/cutadapter:1.0 | gzip > /home/xugang/singularity_image/huawei_file/images/cutadapter.tar.gz
docker save swr.cn-north-4.myhuaweicloud.com/gangxu/fastqc:1.0 | gzip > /home/xugang/singularity_image/huawei_file/images/fastqc.tar.gz
docker save swr.cn-north-4.myhuaweicloud.com/gangxu/base:1.0 | gzip > /home/xugang/singularity_image/huawei_file/images/base.tar.gz
docker save swr.cn-north-4.myhuaweicloud.com/gangxu/ribominer:1.0 | gzip > /Users/xugang/Downloads/data/ribominer.tar.gz
```

