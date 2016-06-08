#!/bin/bash

test -e ssshtest || wget -q https://raw.githubusercontent.com/ryanlayer/ssshtest/master/ssshtest

. ssshtest

set -o nounset


go install -race -a

run check_self_number vcfanno -base-path tests/data/ -lua example/custom.lua tests/data/number.conf tests/data/number-input.vcf
assert_equal 0 $(grep -c "lua error in postannotation" $STDERR_FILE)
cat $STDERR_FILE

run check_example vcfanno -lua example/custom.lua example/conf.toml example/query.vcf.gz
assert_equal $(zgrep -cv ^# example/query.vcf.gz) $(grep -cv ^# $STDOUT_FILE)
assert_equal 6 $(grep ^# $STDOUT_FILE | grep -c lua)
# 2 are for chromsome 2 not found in header
# 1 is for 2:98688 (bedtools intersect -v -a example/query.vcf.gz -b example/fitcons.bed.gz)
# so lua_start doesn't exist.
assert_equal 3 $(grep -c "not found in" $STDERR_FILE)

run check_ends vcfanno -ends -lua example/custom.lua example/conf.toml example/query.vcf.gz
n=$(grep -v ^# $STDOUT_FILE | grep -c right_)
assert_equal $(( $n > 0 )) 1

n=$(grep -v ^# $STDOUT_FILE | grep -c left_)
assert_equal $(( $n > 0 )) 1

n=$(grep  ^"#CHROM" $STDOUT_FILE | cut -f 10- | wc -w) 
assert_equal $n 3

# test bam stuff
n=$(grep -w 10712 $STDOUT_FILE | grep -c "coverage=11;")
assert_equal $n 1


n=$(grep  -v "^##" $STDOUT_FILE | awk 'BEGIN{FS="\t"}{ print NF}' | uniq)
assert_equal $n 12

run check_lua_required vcfanno example/conf.toml example/query.vcf.gz
assert_exit_code 1
assert_in_stderr ERROR
assert_in_stderr lua
assert_no_stdout

cat << EOF > __t.conf
[[annotation]]
file="example/exac.vcf.gz"
ops=["lua:non_existent_field"]
names=["nef"]
fields=["xxx"]
EOF

run check_field_warning vcfanno -lua example/custom.lua __t.conf example/exac.vcf.gz
assert_exit_code 0
assert_in_stderr "xxx not found in header"


cat << EOF > __t.conf
[[postannotation]]
file="example/exac.vcf.gz"
op="lua:non_existent_field"
name="nef"
fields=["xxx"]
EOF

run check_postannotation_field_warning vcfanno -lua example/custom.lua __t.conf example/exac.vcf.gz
assert_exit_code 1
assert_in_stderr "must specify a type"

echo "type='String'" >> __t.conf
run check_postannotation_field_warning vcfanno -lua example/custom.lua __t.conf example/exac.vcf.gz
assert_exit_code 0
assert_in_stderr "xxx not found in header"



run check_cipos vcfanno -base-path tests/citest/ tests/citest/conf.toml  tests/citest/test.vcf  | grep -v ^#
assert_exit_code 0
assert_equal 1 $(grep -c "END=97;ExonGene=GeneY" $STDOUT_FILE)

run check_cipos_ends vcfanno -ends -base-path tests/citest/ tests/citest/conf.toml  tests/citest/test.vcf  | grep -v ^#
assert_exit_code 0
assert_equal 2 $(grep -c "left_ExonGene=GeneY" $STDOUT_FILE)
assert_equal 0 $(grep -c "GeneY,GeneY" $STDOUT_FILE)


run check_multiple_alts vcfanno tests/data/multiple-alts.conf tests/data/multiple-alts.vcf.gz
assert_exit_code 0
# there should be 0 non-header lines without 'max_maf' since we are annotating self.
assert_equal 0 $(grep -v max_maf $STDOUT_FILE | grep -cv ^#)

run check_ends_overlap vcfanno -base-path tests/citest/at/ -ends tests/citest/at/conf.toml tests/citest/at/test.vcf | grep -v ^#
assert_exit_code 0
assert_equal 2 $(grep -c ";left_ExonTranscript=" $STDOUT_FILE)
assert_equal 3 $(grep -c ";right_ExonTranscript=" $STDOUT_FILE)
