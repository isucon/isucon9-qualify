use ExtUtils::MakeMaker;

WriteMakefile(
    NAME          => 'Isucari',
    VERSION_FROM  => 'lib/Isucari.pm',
    PREREQ_PM     => {
        'Kossy' =>  '0.40',
    },
    MIN_PERL_VERSION => '5.008001'
);

