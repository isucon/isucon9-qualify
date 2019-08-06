import * as React from "react";
import {Button} from "@material-ui/core";
import CircularProgress from "@material-ui/core/CircularProgress";

type Props = {
    onClick: (e: React.MouseEvent) => void
    buttonName: string
    loading: boolean
}

export class LoadingButtonComponent extends React.Component<Props> {
    constructor(props: Props) {
        super(props);

        this._onClick = this._onClick.bind(this);
    }

    _onClick(e: React.MouseEvent) {
        e.preventDefault();

        this.props.onClick(e);
    }

    render() {
        const { loading, buttonName } = this.props;

        return (
            <React.Fragment>
                <Button
                    disabled={loading}
                    onClick={this._onClick}
                >
                    {buttonName}
                </Button>
                {loading && <CircularProgress size={24} />}
            </React.Fragment>
        );
    }
};