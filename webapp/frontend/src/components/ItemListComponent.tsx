import React from 'react';
import { ItemData } from '../dataObjects/item';
import makeStyles from '@material-ui/core/styles/makeStyles';
import GridList from '@material-ui/core/GridList';
import { ItemComponent } from './ItemComponent';
import GridListTile from '@material-ui/core/GridListTile';

const useStyles = makeStyles(theme => ({
  grid: {
    width: '300px',
    height: '300px',
  },
}));

interface ItemListPageProps {
  items: ItemData[];
}

const ItemListComponent: React.FC<ItemListPageProps> = ({
  items,
}: ItemListPageProps) => {
  const classes = useStyles();

  const itemComponents = [];

  for (const item of items) {
    itemComponents.push(
      <GridListTile className={classes.grid} key={item.id}>
        <ItemComponent
          itemId={item.id}
          imageUrl={item.thumbnailUrl}
          title={item.name}
          price={item.price}
        />
      </GridListTile>,
    );
  }

  return <GridList cols={3}>{itemComponents}</GridList>;
};

export { ItemListComponent };
