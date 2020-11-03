# Appendix V. Software and Tools

## 1\) Genome Browsers

* [UCSD Genome Browser](https://genome.ucsc.edu/)   \([@youtube](https://youtu.be/eTgEtfI65hA) [@bilibili](https://player.bilibili.com/player.html?aid=30448417&cid=53132461&page=1)\)
* [IGV](http://software.broadinstitute.org/software/igv/)   \([@youtube](https://youtu.be/6_1ZcVw7ptU) [@bilibili](https://player.bilibili.com/player.html?aid=30448472&cid=53133093&page=1)\)

> see more in [our Tutorial](../part-iii.-ngs-data-analyses/1.mapping/1.1-genome-browser.md)

## 2\) DNA-seq

### \(2.1\) Mapping and QC

* **Remove adaptor**: [cutadapt](https://cutadapt.readthedocs.io/en/stable/), [Trimmomatic](http://www.usadellab.org/cms/?page=trimmomatic)
* **Mapping**: [Bowtie2](http://bowtie-bio.sourceforge.net/bowtie2/index.shtml), [STAR](https://github.com/alexdobin/STAR) 
* **QC**: [fastqc](https://www.bioinformatics.babraham.ac.uk/projects/fastqc/)

### \(2.2\) Mutation

* **Mutation discovery**: [GATK](https://gatk.broadinstitute.org/hc/en-us), [Varscan](http://dkoboldt.github.io/varscan/)
* **Mutation annotation**: [ANNOVAR](http://annovar.openbioinformatics.org/en/latest/user-guide/download/)

### \(2.3\) Assembly

* **denovo assembly software**: [Trinity](https://github.com/trinityrnaseq/trinityrnaseq/wiki)

### \(2.4\) CNV

* **Whole Genome Seq**: [Control-FREEC](http://boevalab.inf.ethz.ch/FREEC/) 
* **Whole exome Seq**: [CONTRA](http://contra-cnv.sourceforge.net/), [ExomeCNV](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC3179661/)

### \(2.5\) SV \(structural variation\)

* **structural variation**: [Breakdancer](http://breakdancer.sourceforge.net/)

## 3\) RNA-seq

### \(3.1\) RNA-seq

* **Expression Matrix**: [featureCounts](http://subread.sourceforge.net/), [HTseq](https://htseq.readthedocs.io/en/master/)
* **Differential Expression**: [Deseq2](https://bioconductor.org/packages/release/bioc/html/DESeq2.html), [EdgeR](https://bioconductor.org/packages/release/bioc/html/edgeR.html)
* **Alternative Splicing**: [rMATS](http://rnaseq-mats.sourceforge.net/)
* **RNA Editing**: [RNAEditor](http://rnaeditor.uni-frankfurt.de/), [REDItools](http://code.google.com/p/reditools/)
* ...

### \(3.2\) Single Cell RNA-seq \(scRNA-seq\)

* **Selected  Software providers for scRNA-seq analysis**

> [Nature Biotechnology 2020 38\(3\):254-257](https://www.nature.com/articles/s41587-020-0449-8)

| Software name | Developer | Price structure | Platform-specific | Relevant stages of experiment |
| :--- | :--- | :--- | :--- | :--- |
| [Cell Ranger](https://support.10xgenomics.com/single-cell-gene-expression/software/pipelines/latest/what-is-cell-ranger) | 10X Genomics | Free download | 10X Chromium | Raw read alignment, QC and matrix generation for scRNA-seq and ATAC-seq; data normalization; dimensionality reduction and clustering |
| [Loupe Cell Browser](https://support.10xgenomics.com/single-cell-gene-expression/software/visualization/latest/what-is-loupe-cell-browser) | 10X Genomics | Free download | 10X Chromium | Visualization and analysis |
| [Partek Flow](https://www.partek.com/application-page/single-cell-gene-expression/) | Partek | License | No | Complete data analysis and visualization pipeline for scRNA-seq data |
| [Qlucore Omics Explorer](https://www.qlucore.com/single-cell-rnaseq) | Qlucore | License | No | scRNA-seq data filtering, dimensionality reduction and clustering, visualization |
| [mappa Analysis Pipeline](https://www.takarabio.com/products/automation-systems/icell8-system-and-software/bioinformatics-tools/mappa-analysis-pipeline) | Takara Bio | Free download | Takara ICell8 | Raw read alignment and matrix generation for scRNA-seq |
| [hanta R kit](https://www.takarabio.com/products/automation-systems/icell8-system-and-software/bioinformatics-tools/hanta-r-kit) | Takara Bio | Free download | Takara ICell8 | Clustering and analysis of mappa data |
| [Singular Analysis Toolset](https://www.fluidigm.com/software) | Fluidigm | Free download | Fluidigm C1 or Biomark | Analysis and visualization of differential gene expression data for scRNA-seq |
| [SeqGeq](https://www.flowjo.com/solutions/seqgeq) | FlowJo/BD Biosciences | License | No | Data normalization and QC, dimensionality reduction and clustering, analysis and visualization |
| [Seven Bridges](https://www.sevenbridges.com/bdgenomics/) | Seven Bridges/BD Biosciences | License | BD Rhapsody and Precise | Cloud-based raw read alignment, QC and matrix generation |
| [Tapestri Pipeline/Insights](https://missionbio.com/panels/software/) | Mission Bio | Free download | Mission Bio Tapestri | Analysis of single-cell genomics data |
| [BaseSpace SureCell](https://www.illumina.com/products/by-type/informatics-products/basespace-sequence-hub.html) | Illumina | License | Illumina SureCell libraries | Raw read alignment and matrix generation |
| [OmicSoft Array Studio](https://omicsoftdocs.github.io/ArraySuiteDoc/tutorials/scRNAseq/Introduction/) | Qiagen | License | No | Raw read alignment, QC and matrix generation, dimensionality reduction and clustering |

> QC, quality control; ATAC-seq, assay for transposase-accessible chromatin using sequencing.

## 4\) Interactome

### **\(4.1\) ChIP-seq**

### **\(4.2\) CLIP-seq**

* **Pre-process**: [fastqc](https://www.bioinformatics.babraham.ac.uk/projects/fastqc/)
* **Mapping**: [bowtie](http://bowtie-bio.sourceforge.net/index.shtml), [novoalign](http://www.novocraft.com/products/novoalign/)
* **Peak calling**: [Piranha](http://smithlabresearch.org/software/piranha/), [PARalyzer](https://ohlerlab.mdc-berlin.de/software/PARalyzer_85/), [CIMS](https://zhanglab.c2b2.columbia.edu/index.php/CTK_Documentation)

### **\(4.3\) Motif analysis**

**sequence**

1. MEME motif based sequence analysis tools [http://meme-suite.org/](http://meme-suite.org/)
2. HOMER Software for motif discovery and next-gen sequencing analysis [http://homer.ucsd.edu/homer/motif/](http://homer.ucsd.edu/homer/motif/)

**structure**

1. RNApromo Computational prediction of RNA structural motifs involved in post transcriptional regulatory processes [https://genie.weizmann.ac.il/pubs/rnamotifs08/](https://genie.weizmann.ac.il/pubs/rnamotifs08/)
2. GraphProt modeling binding preferences of RNA-binding proteins [http://www.bioinf.uni-freiburg.de/Software/GraphProt/](http://www.bioinf.uni-freiburg.de/Software/GraphProt/)

## 5\) Epigenetic Data

### **\(5.1\) ChIP-seq**

* **Bi-sulfate data**:
  * Review: [Katarzyna Wreczycka, et al. Strategies for analyzing bisulfite sequencing data. Journal of Biotechnology. 2017.](https://www.sciencedirect.com/science/article/pii/S0168165617315936)
  * Mapping: [Bismark](http://www.bioinformatics.babraham.ac.uk/projects/bismark/), [BSMAP](https://github.com/zyndagj/BSMAPz)
  * Differential Methylation Regions \(DMRs\) detection: [methylkit](https://bioconductor.org/packages/release/bioc/html/methylKit.html), [ComMet](https://github.com/yutaka-saito/ComMet)
  * Segmentation of the methylome, Classification of Fully Methylated Regions \(FMRs\), Unmethylated Regions \(UMRs\) and Low-Methylated Regions \(LMRs\): [MethylSeekR](http://www.bioconductor.org/packages/release/bioc/html/MethylSeekR.html)
  * Annotation of DMRs: [genomation](https://bioconductor.org/packages/release/bioc/html/genomation.html), [ChIPpeakAnno](https://www.bioconductor.org/packages/release/bioc/html/ChIPpeakAnno.html)  
  * Web-based service: [WBSA](http://wbsa.big.ac.cn/)
* **IP data**:
  * Overview to CHIP-Seq: [https://github.com/crazyhottommy/ChIP-seq-analysis](https://github.com/crazyhottommy/ChIP-seq-analysis)
  * peak calling: [MACS2](https://github.com/taoliu/MACS/wiki/Advanced:-Call-peaks-using-MACS2-subcommands)
  * Peak annotation: [HOMER annotatePeak](http://homer.ucsd.edu/homer/ngs/annotation.html), [ChIPseeker](http://bioconductor.org/packages/release/bioc/html/ChIPseeker.html)
  * Gene set enrichment analysis for ChIP-seq peaks: [GREAT](http://bejerano.stanford.edu/great/public/html/)

### **\(5.2\) DNAase-seq**

* review : [Yongjing Liu, et al. Brief in Bioinformatics, 2019.](https://academic.oup.com/bib/article-abstract/20/5/1865/5053117?redirectedFrom=fulltext)
* peek calling:  [F-Seq](http://fureylab.web.unc.edu/software/fseq/)
* peek annotation: [ChIPpeakAnno](https://www.bioconductor.org/packages/release/bioc/html/ChIPpeakAnno.html)
* Motif analysis: [MEME-ChIP](http://meme-suite.org/doc/meme-chip.html?man_type=web)

### **\(5.3\) ATAC-seq**

* pipeline recommended by [Harward informatics](https://github.com/harvardinformatics/ATAC-seq)
* peek calling: [MACS2](https://github.com/taoliu/MACS/wiki/Advanced:-Call-peaks-using-MACS2-subcommands)  
* peak annotation: [ChIPseeker](https://bioconductor.org/packages/release/bioc/html/ChIPseeker.html)
* Motif discovery: [HOMER](http://homer.ucsd.edu/homer/introduction/basics.html)

## 6\) Chromatin and Hi-C

## More: Lu Lab shared tools and scripts

* Scripts:  [Lu Lab](https://github.com/lulab/shared_scripts) \| [Zhi J. Lu](https://github.com/urluzhi/scripts) 
* Plots: [Lu Lab](../part-i.-basic-skills/3.r/3.2.plots-with-r.md) \| [Zhi J. Lu](https://github.com/urluzhi/scripts/tree/master/Rscript/R_plot)



## More: Software for the ages

| Software | Purpose | Creators | Key capabilities | Year released | Citationsa |
| :--- | :--- | :--- | :--- | :---: | :---: |
| BLAST | Sequence alignment | Stephen Altschul, Warren Gish, Gene Myers, Webb Miller, David Lipman | First program to provide statistics for sequence alignment, combination of sensitivity and speed | 1990 | 35,617 |
| R | Statistical analyses | Robert Gentleman, Ross Ihaka | Interactive statistical analysis, extendable by packages | 1996 | N/A |
| ImageJ | Image analysis | Wayne Rasband | Flexibility and extensibility | 1997 | N/A |
| Cytoscape | Network visualization and analysis | Trey Ideker _et al_. | Extendable by plugins | 2003 | 2,374 |
| Bioconductor | Analysis of genomic data | Robert Gentleman _et al_. | Built on R, provides tools to enhance reproducibility of research | 2004 | 3,517 |
| Galaxy | Web-based analysis platform | Anton Nekrutenko, James Taylor | Provides easy access to high-performance computing | 2005 | 309b |
| MAQ | Short-read mapping | Heng Li, Richard Durbin | Integrated read mapping and SNP calling, introduced mapping quality scores | 2008 | 1,027 |
| Bowtie | Short-read mapping | Ben Langmead, Cole Trapnell, Mihai Pop, Steven Salzberg | Fast alignment allowing gaps and mismatches based on Burrows-Wheeler Transform | 2009 | 1,871 |
| Tophat | RNA-seq read mapping | Cole Trapnell, Lior Pachter, Steven Salzberg | Discovery of novel splice sites | 2009 | 817 |
| BWA | Short-read mapping | Heng Li, Richard Durbin | Fast alignment allowing gaps and mismatches based on Burrows-Wheeler Transform | 2009 | 1,556 |
| Circos | Data visualization | Martin Krzywinski _et al_. | Compact representation of similarities and differences arising from comparison between genomes | 2009 | 431 |
| SAMtools | Short-read data format and utilities | Heng Li, Richard Durbin | Storage of large nucleotide sequence alignments | 2009 | 1,551 |
| Cufflinks | RNA-seq analysis | Cole Trapnell, Steven Salzberg, Barbara Wold, Lior Pachter | Transcript assembly and quantification | 2010 | 710 |
| IGV | Short-read data visualization | James Robinson _et al_. | Scalability, real-time data exploration | 2011 | 335 |
| N/A, paper not available in Web of Science. |  |  |  |  |  |

> From: [The anatomy of successful computational biology software](https://www.nature.com/articles/nbt.2721)