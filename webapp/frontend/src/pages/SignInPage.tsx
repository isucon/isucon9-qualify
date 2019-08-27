import React from 'react';
import { makeStyles, Theme } from '@material-ui/core';
import SignInFormContainer from '../containers/SignInFormContainer';
import BasePageContainer from '../containers/BasePageContainer';

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    marginTop: theme.spacing(1),
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
  },
}));

type Props = {};

const SignInPage: React.FC<Props> = () => {
  const classes = useStyles();

  return (
    <BasePageContainer>
      <div className={classes.paper}>
        <SignInFormContainer />
      </div>
    </BasePageContainer>
  );
};

export default SignInPage;
