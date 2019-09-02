import React from 'react';
import Card from '@material-ui/core/Card';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { TransactionItem } from '../../dataObjects/item';
import CardMedia from '@material-ui/core/CardMedia/CardMedia';
import CardContent from '@material-ui/core/CardContent/CardContent';
import Typography from '@material-ui/core/Typography/Typography';
import { TransactionLabel } from '../TransactionLabel';
import { Theme } from '@material-ui/core';
import CardActionArea from '@material-ui/core/CardActionArea';

const MAX_ITEM_NAME_LENGTH = 30;

const useStyles = makeStyles((theme: Theme) => ({
  card: {
    display: 'flex',
    justifyContent: 'flex-start',
    flexDirection: 'row',
  },
  detail: {
    display: 'flex',
    flexDirection: 'column',
  },
  itemTitle: {
    paddingLeft: theme.spacing(1),
    paddingRight: theme.spacing(3),
    paddingBottom: theme.spacing(2),
  },
  img: {
    width: '130px',
    height: '130px',
  },
}));

interface Props {
  item: TransactionItem;
  onClickCard: (item: TransactionItem) => void;
}

const TransactionComponent: React.FC<Props> = ({ item, onClickCard }) => {
  const classes = useStyles();
  const onClick = (e: React.MouseEvent) => {
    e.preventDefault();
    onClickCard(item);
  };

  return (
    <Card>
      <CardActionArea className={classes.card} onClick={onClick}>
        <CardMedia
          className={classes.img}
          image={item.thumbnailUrl}
          title={item.name}
        />
        <div className={classes.detail}>
          <CardContent>
            <Typography
              className={classes.itemTitle}
              noWrap
              variant="subtitle1"
            >
              {item.name.length > MAX_ITEM_NAME_LENGTH
                ? item.name.substr(0, MAX_ITEM_NAME_LENGTH) + '...'
                : item.name}
            </Typography>
            <TransactionLabel itemStatus={item.status} />
          </CardContent>
        </div>
      </CardActionArea>
    </Card>
  );
};

export { TransactionComponent };
