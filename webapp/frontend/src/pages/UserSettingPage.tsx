import React from 'react';
import BasePageContainer from '../containers/BasePageContainer';
import { Grid, Theme } from '@material-ui/core';
import Avatar from '@material-ui/core/Avatar';
import Typography from '@material-ui/core/Typography';
import makeStyles from '@material-ui/core/styles/makeStyles';
import Divider from '@material-ui/core/Divider';
import { InternalServerErrorPage } from './error/InternalServerErrorPage';
import SellingButtonComponent from '../containers/SellingButtonContainer';

const useStyles = makeStyles((theme: Theme) => ({
  avatar: {
    margin: theme.spacing(3),
    width: '100px',
    height: '100px',
  },
  divider: {
    margin: theme.spacing(1),
  },
  descSection: {
    marginTop: theme.spacing(3),
    marginBottom: theme.spacing(3),
  },
}));

type Props = {
  id?: number;
  accountName?: string;
  address?: string;
  numSellItems?: number;
};

const UserSettingPage: React.FC<Props> = ({
  id,
  accountName,
  address,
  numSellItems,
}) => {
  const classes = useStyles();

  if (
    id === undefined ||
    accountName === undefined ||
    address === undefined ||
    numSellItems === undefined
  ) {
    return (
      <InternalServerErrorPage message="ユーザ情報が読み込めませんでした" />
    );
  }

  return (
    <BasePageContainer>
      <Grid
        container
        direction="row"
        justify="center"
        alignItems="center"
        wrap="nowrap"
        spacing={2}
      >
        <Grid item xs={3}>
          <Avatar className={classes.avatar}>{accountName.charAt(0)}</Avatar>
        </Grid>
        <Grid item xs={9}>
          <Typography variant="h3">{accountName}</Typography>
        </Grid>
      </Grid>
      <Grid container>
        <Grid item xs={12}>
          <div className={classes.descSection}>
            <Typography variant="h4">住所</Typography>
            <Divider className={classes.divider} variant="middle" />
            <Typography variant="body1">{address}</Typography>
          </div>
          <div className={classes.descSection}>
            <Typography variant="h4">出品数</Typography>
            <Divider className={classes.divider} variant="middle" />
            <Typography variant="body1">{numSellItems}</Typography>
          </div>
        </Grid>
      </Grid>
      <SellingButtonComponent />
    </BasePageContainer>
  );
};

export default UserSettingPage;
