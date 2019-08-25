import React from 'react';
import Card from '@material-ui/core/Card';
import { Link as RouteLink } from 'react-router-dom';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { routes } from '../../routes/Route';
import { TransactionItem } from '../../dataObjects/item';
import CardMedia from '@material-ui/core/CardMedia/CardMedia';
import CardContent from '@material-ui/core/CardContent/CardContent';
import Typography from '@material-ui/core/Typography/Typography';
import { TransactionLabel } from '../TransactionLabelComponent';

const useStyles = makeStyles(theme => ({
  card: {
    display: 'flex',
  },
}));

interface Props {
  item: TransactionItem;
}

const TransactionComponent: React.FC<Props> = ({ item }) => {
  const classes = useStyles();

  return (
    <Card className={classes.card}>
      <RouteLink to={routes.transaction.getPath(item.id)}>
        <Card>
          <CardMedia image={item.thumbnailUrl} title={item.name} />
          <CardContent>
            <Typography>{item.name}</Typography>
            <TransactionLabel itemStatus={item.status} />
          </CardContent>
        </Card>
      </RouteLink>
    </Card>
  );
};

export { TransactionComponent };
