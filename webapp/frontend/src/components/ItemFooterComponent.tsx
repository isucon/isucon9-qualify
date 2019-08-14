import { AppBar } from '@material-ui/core';
import makeStyles from '@material-ui/core/styles/makeStyles';
import Grid from '@material-ui/core/Grid';
import * as React from 'react';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';

const useStyles = makeStyles(theme => ({
  appBar: {
    top: 'auto',
    bottom: 0,
  },
  buyButton: {
    margin: theme.spacing(1),
  },
}));

type Props = {
  price: number;
  onClick: (e: React.MouseEvent) => void;
  buttonText: string;
};

const ItemFooterComponent: React.FC<Props> = ({
  price,
  onClick,
  buttonText,
}) => {
  const classes = useStyles();

  return (
    <AppBar color="primary" position="fixed" className={classes.appBar}>
      <Grid container spacing={2} direction="row" alignItems="center">
        <Grid item>
          <Typography variant="h5">{price}ｲｽｺｲﾝ</Typography>
        </Grid>
        <Grid item>
          <Button
            variant="contained"
            className={classes.buyButton}
            onClick={onClick}
          >
            {buttonText}
          </Button>
        </Grid>
      </Grid>
    </AppBar>
  );
};

export default ItemFooterComponent;
