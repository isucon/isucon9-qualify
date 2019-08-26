import React from 'react';
import { makeStyles, Theme } from '@material-ui/core';
import SellFormContainer from '../containers/SellFormContainer';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import BasePageContainer from '../containers/BasePageContainer';

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    marginTop: theme.spacing(1),
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
  },
}));

type Props = {} & ErrorProps;

const SellPage: React.FC<Props> = () => {
  const classes = useStyles();

  return (
    <BasePageContainer>
      <div className={classes.paper}>
        <SellFormContainer />
      </div>
    </BasePageContainer>
  );
};

export default PageComponentWithError<Props>()(SellPage);
