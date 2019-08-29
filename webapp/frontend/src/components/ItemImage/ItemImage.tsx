import React from 'react';
import Card from '@material-ui/core/Card';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { Theme } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';

const getUserStyles = (width: number) =>
  makeStyles((theme: Theme) => ({
    card: {
      width: `${width}px`,
      position: 'relative',
    },
    itemImage: {
      width: `${width}px`,
      height: 'auto',
    },
    soldOut: {
      position: 'absolute',
      top: 0,
      right: 0,
      zIndex: 1,
      width: 0,
      height: 0,
      borderStyle: 'solid',
      borderWidth: `0 140px 140px 0`,
      borderColor: 'transparent #ff0000 transparent transparent;',
    },
    soldOutText: {
      position: 'absolute',
      top: '35px',
      right: '1px',
      fontWeight: theme.typography.fontWeightBold,
      zIndex: 2,
      transform: 'rotate(45deg);',
      color: theme.palette.primary.contrastText,
    },
  }));

interface Props {
  imageUrl: string;
  title: string;
  isSoldOut: boolean;
  width: number;
}

const ItemImage: React.FC<Props> = ({ imageUrl, title, isSoldOut, width }) => {
  const classes = getUserStyles(width)();

  return (
    <Card className={classes.card}>
      {isSoldOut && (
        <React.Fragment>
          <div className={classes.soldOut} />
          <Typography className={classes.soldOutText} variant="h6">
            SOLD OUT
          </Typography>
        </React.Fragment>
      )}
      <img className={classes.itemImage} src={imageUrl} alt={title} />
    </Card>
  );
};

export { ItemImage };
