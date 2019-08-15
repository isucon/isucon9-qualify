import React from 'react';
import { TimelineItem } from '../dataObjects/item';
import makeStyles from '@material-ui/core/styles/makeStyles';
import GridList from '@material-ui/core/GridList';
import { ItemComponent } from './ItemComponent';
import GridListTile from '@material-ui/core/GridListTile';
import InfiniteScroll from 'react-infinite-scroller';
import CircularProgress from '@material-ui/core/CircularProgress/CircularProgress';

const useStyles = makeStyles(theme => ({
  grid: {
    width: '300px',
    height: '300px',
  },
}));

interface ItemListPageProps {
  items: TimelineItem[];
  hasNext: boolean;
  loadMore: (page: number) => void;
}

const ItemListComponent: React.FC<ItemListPageProps> = function({
  items,
  hasNext,
  loadMore,
}: ItemListPageProps) {
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

  return (
    <InfiniteScroll
      pageStart={0}
      loadMore={loadMore}
      hasMore={hasNext}
      loader={<CircularProgress />}
    >
      <GridList cols={3}>{itemComponents}</GridList>
    </InfiniteScroll>
  );
};

export { ItemListComponent };
