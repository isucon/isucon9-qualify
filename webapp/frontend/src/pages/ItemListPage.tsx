import React from 'react';
import { TimelineItem } from '../dataObjects/item';
import { ItemList } from '../components/ItemList';
import SellingButtonContainer from '../containers/SellingButtonContainer';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import BasePageContainer from '../containers/BasePageContainer';
import { createStyles, Theme, WithStyles } from '@material-ui/core';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';
import LoadingComponent from '../components/LoadingComponent';
import Typography from '@material-ui/core/Typography/Typography';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    root: {
      display: 'flex',
      flexWrap: 'wrap',
      marginTop: theme.spacing(1),
      justifyContent: 'space-around',
      overflow: 'hidden',
    },
  });

interface ItemListPageProps extends WithStyles<typeof styles> {
  loading: boolean;
  load: () => void;
  items: TimelineItem[];
  hasNext: boolean;
  loadMore: (createdAt: number, itemId: number, page: number) => void;
}

type Props = ItemListPageProps & ErrorProps;

class ItemListPage extends React.Component<Props> {
  constructor(props: Props) {
    super(props);

    this.props.load();
  }

  render() {
    const { classes, loading, items, loadMore, hasNext } = this.props;

    const Content: React.FC<{}> = () => {
      if (items.length === 0) {
        return (
          <div className={classes.root}>
            <Typography variant="h5">出品されている商品はありません</Typography>
            <SellingButtonContainer />
          </div>
        );
      }

      const lastItem = items[items.length - 1];
      const loadMoreItems = loadMore.bind(
        null,
        lastItem.createdAt,
        lastItem.id,
      );
      return (
        <div className={classes.root}>
          <ItemList items={items} loadMore={loadMoreItems} hasNext={hasNext} />
          <SellingButtonContainer />
        </div>
      );
    };

    return (
      <BasePageContainer>
        {loading ? <LoadingComponent /> : <Content />}
      </BasePageContainer>
    );
  }
}

export default PageComponentWithError<any>()(withStyles(styles)(ItemListPage));
