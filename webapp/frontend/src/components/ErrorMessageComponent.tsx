import React from 'react';
import Typography from "@material-ui/core/Typography/Typography";

interface ErrorMessageComponentProps {
    errMsg: string[]
}

const ErrorMessageComponent: React.FC<ErrorMessageComponentProps> = ({ errMsg }) => {
    const errors = [];

    for (const error of errors) {
        errors.push(
            <Typography
                variant="h5"
                color="error"
            >
                {error}
            </Typography>
        )
    }

    return (
        <div>
            {errors}
        </div>
    );
};

export { ErrorMessageComponent }